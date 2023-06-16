import { Component, Inject, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { BookmarkModel, ItemType } from 'src/app/shared/models/bookmarks.model';
import { ApiBookmarksService } from 'src/app/shared/service/api.bookmarks.service';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { CreateBookmarkModel } from './create.model';

@Component({
  selector: 'create.dialog',
  templateUrl: 'create.dialog.html',
  styleUrls: ['create.dialog.css'],
})
export class CreateBookmarksDialog implements OnInit {

  bookmark: BookmarkModel
  type: string
  selectedPath: string
  toggleCustomFavicon = false;
  customFavicon = '';
  invertFaviconColor = false;
  faviconID = 'temp/-1';
  tempFaviconId = '';
  service: ApiBookmarksService
  snackBar: MatSnackBar;
  showProgress = false;

  constructor(public dialogRef: MatDialogRef<CreateBookmarksDialog>,
    @Inject(MAT_DIALOG_DATA) public data: CreateBookmarkModel)
  {}

  ngOnInit(): void {
    this.service = this.data.service;
    this.snackBar = this.data.snackBar;

    if (this.data.existingBookmark) {
      this.bookmark = this.data.existingBookmark;
      this.faviconID = this.bookmark.id;
      this.type = this.bookmark.type.toString();
      this.selectedPath = this.bookmark.path;
      this.invertFaviconColor = this.bookmark.invertFaviconColor === 1 ? true : false;
    } else {
      this.bookmark = new BookmarkModel();
      this.bookmark.id = '';
      this.bookmark.url = this.data.url;
      if (this.bookmark.url && this.bookmark.url !== null && this.bookmark.url !== '') {
        try {
          const url = new URL(this.bookmark.url);
          let cleansedUrl = url.hostname;
          if (cleansedUrl && cleansedUrl !== '') {
            cleansedUrl = cleansedUrl.replace('www.', '');
            this.bookmark.displayName = cleansedUrl;
          }
        } catch (ex) {
          console.log('could not get hostname for simplification! ' + ex);
        }
      }
      this.type = ItemType.Node.toString();
      this.selectedPath = this.data.currentPath;
    }
  }

  onSave(): void {
    let itemType = ItemType.Node;
    if (this.type === 'Folder') {
      itemType = ItemType.Folder;
    }

    this.bookmark.type = itemType;
    this.bookmark.path = this.selectedPath;
    this.bookmark.favicon = this.tempFaviconId;
    this.bookmark.invertFaviconColor = this.invertFaviconColor ? 1 : 0;
    this.dialogRef.close({
      result: true,
      model: this.bookmark
    });
  }

  onNoClick(): void {
    this.dialogRef.close({
      result: false
    });
  }

  fetchCustomFaviconImageURL(): void {
    if (this.customFavicon === '') {
      return;
    }
    console.log(this.customFavicon);

    this.showProgress = true;
    this.service.createFaviconFromCustomURL(this.customFavicon).subscribe(
      {
        next: (v) => {
          console.log('got value from favicon creation: ' + v.value);
          this.faviconID = `temp/${v.value}`;
          this.showProgress = false;
          this.tempFaviconId = v.value;
        },
        error: (e) => {
          this.showProgress = false;
          new MessageUtils().showError(this.snackBar, e);
        }
      }
    );
  }

  fetchBaseURLFavicon(): void {
    if (this.bookmark.url === '') {
      return;
    }
    console.log(this.bookmark.url);
    this.showProgress = true;
    this.customFavicon = '';
    this.toggleCustomFavicon = false;
    this.service.createBaseURLFavicon(this.bookmark.url).subscribe(
      {
        next: (v) => {
          console.log('got value from favicon creation: ' + v.value);
          this.faviconID = `temp/${v.value}`;
          this.showProgress = false;
          this.tempFaviconId = v.value;
        },
        error: (e) => {
          this.showProgress = false;
          new MessageUtils().showError(this.snackBar, e);
        }
      }
    );
  }
}
