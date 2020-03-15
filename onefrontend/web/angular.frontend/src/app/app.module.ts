import { HttpClientModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { AppRoutingModule } from './app.module.routes';
import { sharedConfig } from './app.module.shared';


@NgModule({
  bootstrap: sharedConfig.bootstrap,
  declarations: sharedConfig.declarations,
  imports: [
      BrowserModule,
      BrowserAnimationsModule,
      FormsModule,
      HttpClientModule,
      AppRoutingModule,
      ...sharedConfig.imports
  ],
  providers: [
      ...sharedConfig.providers
  ],
  entryComponents: sharedConfig.entryComponents
})
export class AppModule { }
