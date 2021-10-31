import { NgModule } from '@angular/core';
import { RouterModule, Routes, UrlMatchResult, UrlSegment } from '@angular/router';
import { AppModules } from './app.globals';
import { BookmarkHomeComponent } from './components/bookmarks/home/home.component';
import { EditSitesComponent } from './components/login/edit/edit.component';
import { SiteHomeComponent } from './components/login/home/home.component';
import { MyDmsDocumentComponent } from './components/mydms/document/document.component';
import { MyDmsHomeComponent } from './components/mydms/home/home.component';

// custom matcher
// match for all URLs starting with 'bookmarks' and collect the sub-path in
// the variable path
// e.g. /bookmarks => path: /
// e.g. /bookmarks/Folder1/Folders2 => path: /Folder1/Folder2
// e.g. /bookmarks/a/b/c/d/e/f/g => path: /a/b/c/d/e/f/g
export function matchStartAndSubPath ( url: UrlSegment[] ): UrlMatchResult {

  if (url.length === 0) {
    return null;
  }

  if (url[0].path === AppModules.Bookmarks) {
    let path = '/';
    if (url.length > 1) {
      url.forEach((e, i) => {
        if (e.path !== AppModules.Bookmarks) {
          if (!path.endsWith('/')) {
            path += '/';
          }
          path += e.path;
        }
      });
    }
    return {
      consumed: url,
      posParams: {
        path: new UrlSegment(path, {})
      }
    };
  }
  return null;
}

const routes: Routes = [
  { path: '', redirectTo: AppModules.Bookmarks, pathMatch: 'full' },

  // sites
  // ------------------------------------------------------------------------
  { path: AppModules.Sites, component: SiteHomeComponent },
  { path: AppModules.Sites + '/edit', component: EditSitesComponent },

  // bookmarks
  // ------------------------------------------------------------------------
  { path: AppModules.Bookmarks, component: BookmarkHomeComponent },
  // use matcher for bookmark-filesystem
  { matcher: matchStartAndSubPath, component: BookmarkHomeComponent },

  // mydms
  // ------------------------------------------------------------------------
  { path: AppModules.MyDMS, component: MyDmsHomeComponent },
  { path: AppModules.MyDMS + '/document/:id', component: MyDmsDocumentComponent },
  { path: AppModules.MyDMS + '/document', component: MyDmsDocumentComponent },

  { path: '**', redirectTo: 'start', }
];

@NgModule({
  imports: [RouterModule.forRoot(routes, { relativeLinkResolution: 'legacy' })],
  exports: [RouterModule]
  })
export class AppRoutingModule {}

