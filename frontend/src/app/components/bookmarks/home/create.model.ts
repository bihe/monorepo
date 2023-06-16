import { MatSnackBar } from '@angular/material/snack-bar';
import { BookmarkModel } from 'src/app/shared/models/bookmarks.model';
import { ApiBookmarksService } from 'src/app/shared/service/api.bookmarks.service';

export interface CreateBookmarkModel {
  currentPath: string;
  absolutePaths: string[];
  existingBookmark: BookmarkModel;
  url: string;
  service: ApiBookmarksService;
  snackBar: MatSnackBar;
}
