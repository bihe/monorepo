import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AppModules } from './app.globals';
import { MyDmsDocumentComponent } from './components/mydms/document/document.component';
import { MyDmsHomeComponent } from './components/mydms/home/home.component';
import { Bookmarkv2RedirectCompnent, Sitesv2RedirectCompnent } from './components/redirect/redirec.component';

const routes: Routes = [
  { path: '', redirectTo: AppModules.MyDMS, pathMatch: 'full' },

  // bookmarks
  // ------------------------------------------------------------------------
  { path: AppModules.Bookmarks, component: Bookmarkv2RedirectCompnent },

  // sites
  // ------------------------------------------------------------------------
  { path: AppModules.Sites, component: Sitesv2RedirectCompnent },

  // mydms
  // ------------------------------------------------------------------------
  { path: AppModules.MyDMS, component: MyDmsHomeComponent },
  { path: AppModules.MyDMS + '/document/:id', component: MyDmsDocumentComponent },
  { path: AppModules.MyDMS + '/document', component: MyDmsDocumentComponent },

  { path: '**', redirectTo: 'start', }
];

@NgModule({
  imports: [RouterModule.forRoot(routes, {})],
  exports: [RouterModule]
})
export class AppRoutingModule { }

