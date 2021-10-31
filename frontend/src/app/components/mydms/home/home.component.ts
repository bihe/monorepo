import { Component, OnDestroy, OnInit } from '@angular/core';
import { MatSlideToggleChange } from '@angular/material/slide-toggle';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import * as moment from 'moment';
import { debounceTime } from 'rxjs/operators';
import { AppModules } from 'src/app/app.globals';
import { AppInfo, WhoAmI } from 'src/app/shared/models/app.info.model';
import { MyDmsDocument } from 'src/app/shared/models/document.model';
import { ProblemDetail } from 'src/app/shared/models/error.problemdetail';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiMydmsService } from 'src/app/shared/service/api.mydms.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-mydms-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class MyDmsHomeComponent implements OnInit,OnDestroy {

  documents: Array<MyDmsDocument> = new Array<MyDmsDocument>();
  totalEntries = 0;
  shownResults = 0;
  showAmount = false;
  readonly baseApiUrl = environment.apiBaseURL;
  appInfo: AppInfo;
  readonly InitialPageSize: number = 20;
  private pagedDocuments: Array<MyDmsDocument> = null;
  private searchString: string = null;

  // all subscriptions are held in this array, on destroy all active subscriptions are unsubscribed
  subscriptions: any[];

  constructor(
    private state: ApplicationState,
    private service: ApiMydmsService,
    private router: Router,
    private snackBar: MatSnackBar,
    private moduleIndex: ModuleIndex,
    private titleService: Title) {

    this.titleService.setTitle('Documents List');
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.MyDMS));
    this.state.setCurrentModule(AppModules.MyDMS);
    this.state.setRoute(this.router.url);

    this.subscriptions = [];

    this.subscriptions.push(this.state.getShowAmount().subscribe(
        x => {
          this.showAmount = x;
        }
      )
    );

    this.subscriptions.push(this.state.getRequestReload().subscribe(
        reload => {
          if (reload) {
            this.documents = new Array<MyDmsDocument>();
            this.searchDocuments(null, 0);
          }
        }
      )
    );

    this.subscriptions.push(this.state.getMyDmsVersion()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      )
    );

    this.subscriptions.push(this.state.getSearchInput().pipe(
      debounceTime(500))
        .subscribe(x => {
          if (!x) {
            return;
          }
          if (x.module != AppModules.MyDMS) {
            return;
          }

          console.log('Search for mydms term: ' + x.term);
          this.searchString = x.term;
          this.documents = [];
          this.searchDocuments(x.term, 0);
        }
      )
    );
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach(sub => {
      sub.unsubscribe();
    });
  }


  ngOnInit() {
    this.searchDocuments(null, 0);
  }

  showMoreResults() {
    this.searchDocuments(this.searchString, this.shownResults);
  }

  addDocument() {
    this.router.navigate(['/' + AppModules.MyDMS + '/document/-1']);
  }

  onShowAmountChanged(showAmountValue: MatSlideToggleChange) {
    this.state.setShowAmount(showAmountValue.checked);

    // update the model so that angular change-detection will work with this
    let docs: MyDmsDocument[];
    docs = [...this.documents];
    this.documents = [];
    docs.forEach(doc => {
      doc.update = Math.random().toString().substr(2, 8);
    });
    this.documents = docs;
  }

  editDocument(doc: MyDmsDocument) {
    this.router.navigate(['/' + AppModules.MyDMS + '/document/' + doc.id]);
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
          this.state.setProgress(false);
          if (error.status == 401 || error.status == 403 || !error.status) {
            this.state.setWhoAmI(new WhoAmI());
            if (!environment.production) {
              window.location.href='assets/noaccess.dev.html';
              return;
            }
            window.location.href='assets/noaccess.html';
            return;
          }
          const errorDetail: ProblemDetail = error;
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
