export enum TagType {
  Tag = 1,
  Sender
}

export class AutoCompleteModel {
  value: any;
  display: string;
  type: TagType;
}

