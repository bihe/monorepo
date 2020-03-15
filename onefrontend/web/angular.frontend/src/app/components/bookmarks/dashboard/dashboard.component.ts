import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { debounceTime } from 'rxjs/operators';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { BookmarkModel } from 'src/app/shared/models/bookmarks.model';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiBookmarksService } from 'src/app/shared/service/api.bookmarks.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { ErrorMode, Errors } from 'src/app/shared/utils/errors';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-bookmarks-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class BookmarkDashBoardComponent implements OnInit {

  bookmarks: BookmarkModel[] = [];
  filtered: BookmarkModel[] = [];
  isUser: boolean = true;
  isAdmin: boolean = false;
  appInfo: AppInfo;
  readonly MaxDashboardEntries = 45;
  readonly baseApiURL = environment.apiUrlBookmarks;

  constructor(private bookmarksService: ApiBookmarksService,
    private snackBar: MatSnackBar,
    private state: ApplicationState,
    private titleService: Title,
    private moduleIndex: ModuleIndex
  ) {
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.Bookmarks));
  }

  ngOnInit() {
    this.titleService.setTitle('bookmarks.Dashboard');
    this.state.setProgress(true);
    this.bookmarksService.getMostVisitedBookmarks(this.MaxDashboardEntries)
      .subscribe(
        data => {
          console.log(data);
          this.state.setProgress(false);
          if (data.count > 0) {
            this.bookmarks = data.value;
          } else {
            this.bookmarks = [];
          }
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

    this.state.getBookmarksVersion()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      );

    this.state.getSearchInput().pipe(
      debounceTime(300))
      .subscribe(s => {
        if (s !== '' && s != null) {
          console.log('Search for: ' + s);
          // search within the existing entries
          this.filtered = this.bookmarks.filter(x => x.displayName.toLowerCase().indexOf(s.toLowerCase()) > -1);
        } else {
          this.filtered = [];
        }
      });
  }

  get dashboardBookmarks(): BookmarkModel[] {
    if (this.filtered.length == 0) {
      return this.bookmarks;
    }
    return this.filtered;
  }

  get defaultFavicon(): string {
    return 'assets/favicon.ico';
  }

  customFavicon(id: string): string {
    return environment.apiUrlBookmarks + `/api/v1/bookmarks/favicon/${id}`;
  }
}
