import { BookmarkModel } from 'src/app/shared/models/bookmarks.model';

export interface CreateBookmarkModel {
  currentPath: string;
  absolutePaths: string[];
  existingBookmark: BookmarkModel;
  url: string;
}
