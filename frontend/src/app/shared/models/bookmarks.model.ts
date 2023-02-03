export enum ItemType {
  Node = "Node",
  Folder = "Folder"
}

export class BookmarkModel {
  public id: string;
  public position: number;
  public path: string;
  public displayName: string;
  public url: string;
  public sortOrder: number;
  public type: ItemType;
  public created: Date;
  public modified: Date;
  public childCount: number;
  public favicon: string;
  public customFavicon: string;
  public highlight: number;
  public showUrl: string;
}

export class BoomarkSortOrderModel {
  public ids: string[]
  public sortOrder: number[];
}

export class BookmarkPathsModel {
  public paths: string[];
  public count: number;
}
