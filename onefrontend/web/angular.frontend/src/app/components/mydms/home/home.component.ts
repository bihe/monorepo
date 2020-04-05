import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import * as moment from 'moment';
import { debounceTime } from 'rxjs/operators';
import { Globals } from 'src/app/app.globals';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { MyDmsDocument } from 'src/app/shared/models/document.model';
import { ProblemDetail } from 'src/app/shared/models/error.problemdetail';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiMydmsService } from 'src/app/shared/service/api.mydms.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { ErrorMode, Errors } from 'src/app/shared/utils/errors';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-mydms-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class MyDmsHomeComponent implements OnInit {

  documents: Array<MyDmsDocument> = new Array<MyDmsDocument>();
  totalEntries = 0;
  shownResults = 0;
  showAmount = false;
  readonly baseApiUrl = environment.apiUrlMydms;
  appInfo: AppInfo;

  readonly InitialPageSize: number = 20;
  private pagedDocuments: Array<MyDmsDocument> = null;
  private searchString: string = null;

  constructor(
    private state: ApplicationState,
    private service: ApiMydmsService,
    private router: Router,
    private snackBar: MatSnackBar,
    private moduleIndex: ModuleIndex,
    private titleService: Title) {

    this.titleService.setTitle('Documents List');
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.MyDMS));

    this.state.getSearchInput().pipe(
      debounceTime(300))
      .subscribe(x => {
        console.log('Search for: ' + x);
        this.searchString = x;
        this.documents = [];
        this.searchDocuments(x, 0);
      });

    this.state.getShowAmount().subscribe(
      x => {
        this.showAmount = x;
      }
    );

    this.state.getRequestReload().subscribe(
      reload => {
        if (reload) {
          this.documents = new Array<MyDmsDocument>();
          this.searchDocuments(null, 0);
        }
      }
    );

    this.state.getMyDmsVersion()
    .subscribe(
      x => {
        this.appInfo = x;
      }
    );

    this.state.setRoute(this.router.url);
  }

  ngOnInit() {
    this.searchDocuments(null, 0);
  }

  showMoreResults() {
    this.searchDocuments(this.searchString, this.shownResults);
  }

  clearSearch() {
    this.state.setSearchInput('');
  }

  addDocument() {
    this.router.navigate(['/' + Globals.MyDmsPath + '/document/-1']);
  }

  editDocument(doc: MyDmsDocument) {
    this.router.navigate(['/' + Globals.MyDmsPath + '/document/' + doc.id]);
  }

  searchDocuments(title: string, skipEntries: number) {
    this.state.setProgress(true);
    this.service.searchDocuments(title, this.InitialPageSize, skipEntries)
      .subscribe(
        result => {
          const returnedResults = (result.documents) ? result.documents.length : 0;
          this.totalEntries = result.totalEntries;
          this.shownResults = skipEntries + returnedResults;

          const doucmentResult = result.documents;
          console.log('Result from search: ' + returnedResults);
          if (doucmentResult) {
            this.pagedDocuments = new Array<MyDmsDocument>();
            doucmentResult.forEach(a => {
              const doc = new MyDmsDocument();
              doc.title = a.title;
              doc.created = a.created;
              doc.modified = a.modified;
              if (this.showAmount) {
                doc.amount = a.amount;
              }
              doc.fileName = a.fileName;
              doc.encodedFilename = btoa(encodeURI(a.fileName));
              doc.previewLink = a.previewLink;
              doc.id = a.id;
              doc.tags = a.tags;
              doc.senders = a.senders;
              doc.dateHuman = moment(doc.lastDate).fromNow();
              doc.invoiceNumber = a.invoiceNumber;

              this.pagedDocuments.push(doc);
            });
            this.documents = this.documents.concat(this.pagedDocuments);
            this.documents = arrayUnique(this.documents);
          }
          this.state.setProgress(false);
        },
        error => {
          if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
            window.location.reload();
            return;
          }

          const errorDetail: ProblemDetail = error;
          this.state.setProgress(false);
          console.log(errorDetail);
          new MessageUtils().showError(this.snackBar, errorDetail.title);
        }
      );
  }
}

function arrayUnique(array: any) {
  const a = array.concat();
  for (let i = 0; i < a.length; ++i) {
    for (let j = i + 1; j < a.length; ++j) {
      if (a[i].id === a[j].id) {
        a.splice(j--, 1);
      }
    }
  }
  return a;
}
