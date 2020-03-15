import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';
import { Component, OnInit } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { ActivatedRoute, Router } from '@angular/router';
import { debounceTime, flatMap, map, switchMap } from 'rxjs/operators';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { BookmarkModel, BoomarkSortOrderModel } from 'src/app/shared/models/bookmarks.model';
import { ProblemDetail } from 'src/app/shared/models/error.problemdetail';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiBookmarksService } from 'src/app/shared/service/api.bookmarks.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { ErrorMode, Errors } from 'src/app/shared/utils/errors';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';
import { ConfirmDialogComponent, ConfirmDialogModel } from '../../confirm-dialog/confirm-dialog.component';
import { CreateBookmarksDialog } from './create.dialog';

const bookmarkPath = '/bookmarks'

@Component({
  selector: 'app-bookmarks-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class BookmarkHomeComponent implements OnInit {

  bookmarks: BookmarkModel[] = [];
  currentPath: string = '';
  pathElemets: string[] = [];
  absolutePaths: string[] = [];
  isUser: boolean = true;
  isAdmin: boolean = false;
  search: string = '';
  searchMode: boolean = false;
  changePath: boolean = false;
  pathInput: string = '';
  highlightDropZone: boolean = false;
  appInfo: AppInfo;
  readonly baseApiURL = environment.apiUrlBookmarks;

  constructor(private bookmarksService: ApiBookmarksService,
    private snackBar: MatSnackBar,
    public dialog: MatDialog,
    private state: ApplicationState,
    private router: Router,
    private activeRoute: ActivatedRoute,
    private titleService: Title,
    private moduleIndex: ModuleIndex
  ) {
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.Bookmarks));
  }

  ngOnInit() {
    this.activeRoute.params
      .pipe(
        map(p => {
          if (p.path) {
            // we need to have an "absolute" path
            let path = p.path;
            if (!path.startsWith('/')) {
              path = "/" + path;
            }
            return path;
          }
          return '/';
        }),
        flatMap(path => {
          this.state.setProgress(true);
          return this.bookmarksService.getBookmarkFolderByPath(path);
        }),
        switchMap(folderResult => {
          // if we have received a folder, we need to get the path!
          if (folderResult.success === true) {
            if (folderResult.value.displayName === 'Root') {
              this.titleService.setTitle('Bookmarks');
            } else {
              this.titleService.setTitle(folderResult.value.displayName);
            }

            let path = folderResult.value.path;
            if (!path.endsWith('/')) {
              path += '/';
            }
            // special root treatment!
            if (folderResult.value.displayName !== 'Root') {
              path = path + folderResult.value.displayName;
            }
            this.currentPath = path;
            this.pathElemets = this.splitPathElements(path);
            return this.bookmarksService.getBookmarksForPath(path)
          }
        })
      )
      .subscribe(
        data => {
          console.log(data);
          this.state.setProgress(false);
          if (data.count > 0) {
            this.bookmarks = data.value;
          } else {
            this.bookmarks = [];
          }
          console.log('currentPath: ' + this.currentPath);
        },
        error => {
          if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
            window.location.reload();
            return;
          }
          this.state.setProgress(false);
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error.title);
        }
      );

    this.state.isAdmin().subscribe(
      data => {
        this.isAdmin = data;
      }
    );

    this.state.getSearchInput().pipe(
      debounceTime(300))
      .subscribe(x => {
        if (x !== '' && x != null) {
          console.log('Search for: ' + x);
          this.searchBookmarks(x);
        } else {
          this.getBookmarksForPath(this.currentPath);
        }
      });

    this.state.getBookmarksVersion()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      );
  }

  gotoPath(path: string) {
    if (path.startsWith('//')) {
      path = path.replace('//', '/'); // fix for the root path!
    }
    console.log('goto: ' + path);
    if (path === this.currentPath) {
      // reload the data
      this.getBookmarksForPath(path);
    }
    else {
      // push the path to the URL - pushstate like
      this.router.navigate([bookmarkPath + path]);
      this.searchMode = false;
      if (this.changePath === true) {
        this.pathInput = path;
      }
    }
  }

  getBookmarksForPath(path: string) {
    this.searchMode = false;
    this.state.setProgress(true);
    this.bookmarksService.getBookmarksForPath(path)
      .subscribe(
        data => {
          this.state.setProgress(false);
          if (data.count > 0) {
            this.bookmarks = data.value;
          } else {
            this.bookmarks = [];
          }
          this.currentPath = path;
          this.pathElemets = this.splitPathElements(path);
          console.log('currentPath: ' + this.currentPath);
        },
        error => {
          if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
            window.location.reload();
            return;
          }
          this.state.setProgress(false);
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error.title);
        }
      );
  }

  editBookmark(id: string) {
    console.log('edit bookmark: ' + id);
    this.bookmarksService.fetchBookmarkById(id).subscribe(
      data => {
        this.bookmarksService.getAllPaths().subscribe(
          allPaths => {
            this.state.setProgress(false);
            console.log(allPaths);

            const dialogRef = this.dialog.open(CreateBookmarksDialog, {
              panelClass: 'my-full-screen-dialog',
              data: {
                absolutePaths: allPaths.paths,
                currentPath: this.currentPath,
                existingBookmark: data
              }
            });

            dialogRef.afterClosed().subscribe(data => {
              console.log('dialog was closed');
              if (data.result) {
                let bookmark: BookmarkModel = data.model;
                if (typeof bookmark.favicon === 'undefined' || bookmark.favicon === null) {
                  bookmark.favicon = '';
                }
                console.log(bookmark);

                // update the UI immediately!
                let oldDisplayName = '';
                let existing = this.bookmarks.find(x => x.id === bookmark.id);
                if (existing) {
                  oldDisplayName = existing.displayName;
                  existing.displayName = bookmark.displayName;
                }

                this.bookmarksService.updateBookmark(bookmark).subscribe(
                  data => {
                    this.state.setProgress(false);
                    console.log(data);
                    if (data.success) {
                      new MessageUtils().showSuccess(this.snackBar, data.message);
                      this.getBookmarksForPath(this.currentPath);
                    }
                  },
                  error => {
                    if (oldDisplayName && existing) {
                      existing.displayName = oldDisplayName;
                    }
                    if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
                      window.location.reload();
                      return;
                    }
                    const errorDetail: ProblemDetail = error;
                    this.state.setProgress(false);
                    console.log(errorDetail);
                    new MessageUtils().showError(this.snackBar, errorDetail.title);
                  }
                );
              }
            });
          },
          error => {
            if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
              window.location.reload();
              return;
            }
            const errorDetail: ProblemDetail = error;
            this.state.setProgress(false);
            console.log(errorDetail);
            new MessageUtils().showError(this.snackBar, errorDetail.title);
          }
        );
      },
      error => {
        if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
          window.location.reload();
          return;
        }
        const errorDetail: ProblemDetail = error;
        this.state.setProgress(false);
        console.log(errorDetail);
        new MessageUtils().showError(this.snackBar, errorDetail.title);
      }
    );
  }

  addBookmark(url: string) {
    console.log('add bookmark!');
    const dialogRef = this.dialog.open(CreateBookmarksDialog, {
      panelClass: 'my-full-screen-dialog',
      data: {
        absolutePaths: this.absolutePaths,
        currentPath: this.currentPath,
        url: url
      }
    });

    dialogRef.afterClosed().subscribe(data => {
      console.log('dialog was closed');
      if (data.result) {
        let bookmark: BookmarkModel = data.model;
        bookmark.favicon = '';
        console.log(bookmark);

        this.bookmarksService.createBookmark(bookmark).subscribe(
          data => {
            this.state.setProgress(false);
            console.log(data);
            if (data.success) {
              new MessageUtils().showSuccess(this.snackBar, data.message);
              this.getBookmarksForPath(this.currentPath);
            }
          },
          error => {
            if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
              window.location.reload();
              return;
            }
            const errorDetail: ProblemDetail = error;
            this.state.setProgress(false);
            console.log(errorDetail);
            new MessageUtils().showError(this.snackBar, errorDetail.title);
          }
        );
      }
    });
  }

  deleteBookmark(id: string) {
    const dialogData = new ConfirmDialogModel('Confirm delete!', 'Really delete bookmark?');

    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      maxWidth: "400px",
      data: dialogData
    });

    dialogRef.afterClosed().subscribe(dialogResult => {
      console.log(dialogResult);
      if (dialogResult === true) {
        console.log('Will delete bookmark by id ' + id);

        this.bookmarksService.deleteBookmarkById(id).subscribe(
          data => {
            this.state.setProgress(false);
            console.log(data);
            if (data.success) {
              new MessageUtils().showSuccess(this.snackBar, data.message);
              this.getBookmarksForPath(this.currentPath);
            }
          },
          error => {
            if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
              window.location.reload();
              return;
            }
            const errorDetail: ProblemDetail = error;
            this.state.setProgress(false);
            console.log(errorDetail);
            new MessageUtils().showError(this.snackBar, errorDetail.title);
          }
        );
      }
    });
  }

  searchBookmarks(search: string) {
    console.log('search for ' + search);
    this.state.setProgress(true);
    this.bookmarksService.getBookmarksByName(search)
      .subscribe(
        data => {
          this.state.setProgress(false);
          if (data.count > 0) {
            this.bookmarks = data.value;
            this.searchMode = true;
            this.currentPath = '/';
            this.pathElemets = this.splitPathElements(this.currentPath);
          } else {
            this.bookmarks = [];
          }
        },
        error => {
          this.state.setProgress(false);
          if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
            window.location.reload();
            return;
          }
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error.title);
        }
      );
  }

  editMode(active: boolean) {
    this.changePath = active;
    if (active) {
      this.pathInput = this.currentPath;
    } else {
      this.pathInput = '';
    }
  }

  doChangePath() {
    if (this.changePath === true && this.pathInput !== '') {
      this.gotoPath(this.pathInput);
    } else {
      console.log('cannot change path - invalid!');
    }
  }

  drop(event: CdkDragDrop<string[]>) {
    if (event.previousIndex == event.currentIndex) {
      return;
    }

    console.log(`move items in list from ${event.previousIndex} to ${event.currentIndex}.`);
    let selectedItem = this.bookmarks[event.previousIndex];
    let targetItem = this.bookmarks[event.currentIndex];

    // get the sortOrder of the target-element
    let sortOrder = targetItem.sortOrder;

    if (event.currentIndex > event.previousIndex) {
      // put it "after" the selected item
      sortOrder += 1;
      selectedItem.sortOrder = sortOrder;
    } else {
      // we want to put the item "in front" of this item
      sortOrder -= 1;
      selectedItem.sortOrder = sortOrder;
    }

    // we have the new sort-order of the selected item
    // update the sort-order of the item in the backend

    // swap in UI (faster)
    let oldBookmarkList = [...this.bookmarks]; // clone the array
    moveItemInArray(this.bookmarks, event.previousIndex, event.currentIndex);

    const sortOrderModel = new BoomarkSortOrderModel();
    sortOrderModel.ids = [];
    sortOrderModel.sortOrder = [];
    this.bookmarks.forEach((e, index) => {
      sortOrderModel.ids.push(e.id);
      sortOrderModel.sortOrder.push(index);
    });

    this.bookmarksService.updateBookmarksSortOrder(sortOrderModel).subscribe(
      data => {
        this.state.setProgress(false);
        if (!data.success) {
          console.log('could not update sortOrder server-side - restore to previous');
          this.bookmarks = oldBookmarkList;
          new MessageUtils().showError(this.snackBar, 'could not update the bookmarks sort-order!');
        }
      },
      error => {
        if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
          window.location.reload();
          return;
        }
        // an error occured - we could not update the bookmarks server-side
        // restore the old behavior
        this.bookmarks = oldBookmarkList;
        const errorDetail: ProblemDetail = error;
        this.state.setProgress(false);
        console.log(errorDetail);
        new MessageUtils().showError(this.snackBar, errorDetail.title);
      }
    );

  }

  get defaultFavicon(): string {
    return 'assets/favicon.ico';
  }

  customFavicon(id: string): string {
    return environment.apiUrlBookmarks + `/api/v1/bookmarks/favicon/${id}`;
  }

  dragEnter(ev: any, highlight: boolean) {
    ev.preventDefault();
    this.highlightDropZone = highlight;
  }

  doDropText(ev: any) {
    ev.preventDefault();
    this.highlightDropZone = false;
    let url = ev.dataTransfer.getData('text');
    if (url) {
      console.log(`url ${url} dropped!`);
      if (!url.startsWith('http')) {
        url = 'https://' + url; // chrome does not willingly provide the scheme
      }
      this.addBookmark(url);
    }
  }

  private splitPathElements(path: string): string[] {
    let parts = path.split('/');
    if (parts && parts.length > 0) {
      if (parts[0] === '') {
        parts[0] = '/';
      }
    } else {
      parts = [];
      parts[0] = '/';
    }

    if (parts[parts.length - 1] === '') {
      parts.pop();
    }

    // also create a list of the path-elements which create the absolute path
    // for each element
    this.absolutePaths = [];
    let absPath = '';
    parts.forEach(e => {
      if (absPath !== '' && !absPath.endsWith('/')) {
        absPath += '/';
      }
      absPath += e;
      this.absolutePaths.push(absPath);
    });

    return parts;
  }
}
