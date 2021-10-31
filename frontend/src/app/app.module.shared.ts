import { DragDropModule } from '@angular/cdk/drag-drop';
import { NgModule } from '@angular/core';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatBadgeModule } from '@angular/material/badge';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatOptionModule } from '@angular/material/core';
import { MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatRadioModule } from '@angular/material/radio';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Title } from '@angular/platform-browser';
import { LazyLoadImageModule } from 'ng-lazyload-image';
import { AppComponent } from './components/app/app.component';
import { CreateBookmarksDialog } from './components/bookmarks/home/create.dialog';
import { BookmarkHomeComponent } from './components/bookmarks/home/home.component';
import { ConfirmDeleteDialogComponent } from './components/confirm-delete-dialog/confirmation.component';
import { ConfirmDialogComponent } from './components/confirm-dialog/confirm-dialog.component';
import { HeaderComponent } from './components/header/header.component';
import { EditSitesComponent } from './components/login/edit/edit.component';
import { SiteHomeComponent } from './components/login/home/home.component';
import { MyDmsDocumentComponent } from './components/mydms/document/document.component';
import { MyDmsHomeComponent } from './components/mydms/home/home.component';
import { ModuleIndex } from './shared/moduleIndex';
import { DateFormatPipe } from './shared/pipes/dataformat';
import { EllipsisPipe } from './shared/pipes/ellipsis';
import { ApiBookmarksService } from './shared/service/api.bookmarks.service';
import { ApiCoreService } from './shared/service/api.core.service';
import { ApiMydmsService } from './shared/service/api.mydms.service';
import { ApiSiteService } from './shared/service/api.sites.service';
import { ApplicationState } from './shared/service/application.state';

@NgModule({
  imports: [ MatProgressSpinnerModule, MatTooltipModule, MatSnackBarModule, MatButtonModule, MatDialogModule, MatInputModule, MatFormFieldModule, MatRadioModule, MatOptionModule, MatSelectModule, MatMenuModule, MatIconModule, MatBadgeModule, DragDropModule, MatCheckboxModule, MatCardModule, MatChipsModule, MatSlideToggleModule, MatExpansionModule, MatChipsModule, MatAutocompleteModule, MatProgressBarModule ],
  exports: [ MatProgressSpinnerModule, MatTooltipModule, MatSnackBarModule, MatButtonModule, MatDialogModule, MatInputModule, MatFormFieldModule, MatRadioModule, MatOptionModule, MatSelectModule, MatMenuModule, MatIconModule, MatBadgeModule, DragDropModule, MatCheckboxModule, MatCardModule, MatChipsModule, MatSlideToggleModule, MatExpansionModule, MatChipsModule, MatAutocompleteModule, MatProgressBarModule ],
})
export class AppMaterialModule { }

export const sharedConfig: NgModule = {
    bootstrap: [ AppComponent ],
    declarations: [
      AppComponent,
      BookmarkHomeComponent,
      HeaderComponent,
      EllipsisPipe,
      DateFormatPipe,
      CreateBookmarksDialog,
      ConfirmDialogComponent,
      MyDmsHomeComponent,
      SiteHomeComponent,
      EditSitesComponent,
      ConfirmDeleteDialogComponent,
      MyDmsDocumentComponent
    ],
    imports: [
      AppMaterialModule,
      LazyLoadImageModule
    ],
    providers: [
      ApplicationState,
      ApiBookmarksService,
      ApiSiteService,
      ApiMydmsService,
      ApiCoreService,
      Title,
      ModuleIndex
    ],
    entryComponents: [ CreateBookmarksDialog, ConfirmDialogComponent ]
};
