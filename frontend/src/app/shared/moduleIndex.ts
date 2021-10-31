import { ModuleInfo } from './models/module.model';
import { Injectable } from "@angular/core";

export enum ModuleName {
  Bookmarks,
  Sites,
  MyDMS
}

@Injectable()
export class ModuleIndex {
  private modules: Map<ModuleName, ModuleInfo>

  constructor() {
    this.modules = new Map<ModuleName, ModuleInfo>();
    this.modules[ModuleName.Bookmarks] = Object.assign(new ModuleInfo(), {
      displayName: 'bookmarks',
      classValue: 'fa fa-bookmark logo'
    });

    this.modules[ModuleName.Sites] = Object.assign(new ModuleInfo(), {
      displayName: 'sites',
      classValue: 'fa fa-shield logo'
    });

    this.modules[ModuleName.MyDMS] = Object.assign(new ModuleInfo(), {
      displayName: 'mydms',
      classValue: 'fa fa-file logo'
    });
  }

  public getModuleInfo(module: ModuleName): ModuleInfo {
    return this.modules[module];
  }
}
