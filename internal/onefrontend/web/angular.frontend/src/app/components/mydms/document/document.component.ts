import { COMMA, ENTER } from '@angular/cdk/keycodes';
import { Component, ElementRef, EventEmitter, OnInit, ViewChild } from '@angular/core';
import { FormControl } from '@angular/forms';
import { MatAutocomplete, MatAutocompleteSelectedEvent } from '@angular/material/autocomplete';
import { MatChipInputEvent } from '@angular/material/chips';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { ActivatedRoute, Router } from '@angular/router';
import { humanizeBytes, UploadFile, UploadInput, UploadOutput } from 'ngx-uploader';
import { Globals } from 'src/app/app.globals';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { MyDmsDocument } from 'src/app/shared/models/document.model';
import { ApiMydmsService } from 'src/app/shared/service/api.mydms.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';
import { ConfirmDeleteDialogComponent } from '../../confirm-delete-dialog/confirmation.component';

@Component({
  selector: 'app-mydms-document',
  templateUrl: './document.component.html',
  styleUrls: ['./document.component.css']
})
export class MyDmsDocumentComponent implements OnInit {

  isNewDocument = true;
  document: MyDmsDocument = new MyDmsDocument();

  selectedTags: any[] = [];
  selectedSenders: any[] = [];
  senders: string[];
  tags: string[];

  files: UploadFile[];
  uploadInput: EventEmitter<UploadInput>;
  humanizeBytes: Function;
  dragOver: boolean;

  documentTitle = '';
  documentAmount = 0;
  invoiceNumber = '';
  uploadFileName = '';
  readonly baseApiUrl = environment.apiUrlMydms;
  appInfo: AppInfo;

  encodedUploadFileName = '';
  showAmount = false;

  // settings for encryption
  encPassword = '';
  initPassword = '';

  // new chips
  visible = true;
  selectable = true;
  removable = true;
  separatorKeysCodes: number[] = [ENTER, COMMA];

  tagCtrl = new FormControl();
  filteredTags: string[];
  @ViewChild('tagInput') tagInput: ElementRef<HTMLInputElement>;
  @ViewChild('autoTag') matAutocompleteTag: MatAutocomplete;

  senderCtrl = new FormControl();
  filteredSenders: string[];
  @ViewChild('senderInput') senderInput: ElementRef<HTMLInputElement>;
  @ViewChild('autoSender') matAutocompleteSender: MatAutocomplete;


  private uploadToken = '';

  constructor(
    private service: ApiMydmsService,
    private state: ApplicationState,
    private snackBar: MatSnackBar,
    private route: ActivatedRoute,
    private dialog: MatDialog,
    private router: Router,
    private titleService: Title) {
    this.files = []; // local uploading files array
    this.uploadInput = new EventEmitter<UploadInput>(); // input events, we use this to emit data to ngx-uploader
    this.humanizeBytes = humanizeBytes;

      this.senderCtrl.valueChanges.subscribe(
        item => {
          this.service.searchSenders(item).subscribe(
            v => {
               this.filteredSenders = v.result;
             }
           );
        }
      );

      this.tagCtrl.valueChanges.subscribe(
        item => {
          this.service.searchTags(item).subscribe(
            v => {
               this.filteredTags = v.result;
             }
           );
        }
      );
  }

  ngOnInit() {

    this.state.getMyDmsVersion()
    .subscribe(
      x => {
        this.appInfo = x;
      }
    );
    this.isNewDocument = true;

    this.state.getShowAmount().subscribe(
      x => {
        this.showAmount = x;
      }
    );

    this.state.setRoute(this.router.url);

    const id = this.route.snapshot.params['id'] || -1;
    console.log('Got route id: ' + id);
    if (id === -1 || id === '-1') {
      this.document = null;
      this.titleService.setTitle('New Document');
      return;
    }

    this.state.setProgress(true);
    this.service.getDocument(id)
      .subscribe(
        result => {
          if (result) {
            this.document = result;
            if (this.document.senders && this.document.senders.length > 0) {
              this.senders = this.document.senders;
            }

            if (this.document.tags && this.document.tags.length > 0) {
              this.tags = this.document.tags;
            }
            this.isNewDocument = false;

            this.documentTitle = this.document.title;
            if (this.showAmount) {
              this.documentAmount = this.document.amount;
            }
            this.uploadFileName = this.document.fileName;
            this.uploadToken = '-';
            this.invoiceNumber = this.document.invoiceNumber;
            this.encodedUploadFileName = this.document.previewLink;

            if (this.document.senders) {
              this.document.senders.forEach(item => {
                const sender: any = {};
                sender.value = item;
                sender.display = item;
                this.selectedSenders.push(item);
              });
            }

            if (this.document.tags) {
              this.document.tags.forEach(item => {
                const tag: any = {};
                tag.value = item;
                tag.display = item;
                this.selectedTags.push(item);
              });
            }
            this.titleService.setTitle('Document: ' + this.document.title);
            this.state.setProgress(false);
          }
        },
        error => {
          this.state.setProgress(false);
          new MessageUtils().showError(this.snackBar, error);
        }
      );


  }

  public onCancel() {
    this.router.navigate(['/' + Globals.MyDmsPath]);
  }

  public onSave() {
    if (this.isFormValid()) {
      //this.convertSenderAndTags();
      this.tags = this.selectedTags;
      this.senders = this.selectedSenders;

      if (this.document === null) {
        this.document = new MyDmsDocument();
      }

      this.document.title = this.documentTitle;
      this.document.amount = this.documentAmount;
      this.document.senders = this.senders;
      this.document.tags = this.tags;
      this.document.uploadFileToken = this.uploadToken;
      this.document.fileName = this.uploadFileName;
      this.document.invoiceNumber = this.invoiceNumber;

      this.state.setProgress(true);
      this.service.saveDocument(this.document)
        .subscribe(
          result => {
            if (result) {
              if (result.result === 'saved') {
                this.state.setProgress(false);
                console.log(result.message);
                this.router.navigate(['/' + Globals.MyDmsPath]);
                return;
              } else {
                this.state.setProgress(false);
                new MessageUtils().showError(this.snackBar, result.message);
              }
            }
          },
          error => {
            this.state.setProgress(false);
            console.error(error.detail);
            new MessageUtils().showError(this.snackBar, error.detail);
          }
        );
    } else {
      new MessageUtils().showError(this.snackBar, 'The form is not valid!');
    }
  }

  public onDelete() {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, ConfirmDeleteDialogComponent.getDialogConfig(this.documentTitle));

    dialogRef.afterClosed().subscribe(result => {
      if (result === true) {
        console.log('Delete confirmed!');
        this.state.setProgress(true);
        this.service.deleteDocument(this.document.id)
          .subscribe(
            r => {
              if (r.result === 'deleted') {
                this.state.setProgress(false);
                console.log(r.message);
                this.router.navigate(['/' + Globals.MyDmsPath]);
                return;
              } else {
                this.state.setProgress(false);
                new MessageUtils().showError(this.snackBar, result.message);
              }
            },
            error => {
              this.state.setProgress(false);
              new MessageUtils().showError(this.snackBar, error);
            }
          );

      }
    });
  }

  public onUploadOutput(output: UploadOutput): void {
    if (output.type === 'allAddedToQueue') { // when all files added in queue
      this.state.setProgress(true);
      // console.log(`passwords: '${this.encPassword}'; initialPassword: '${this.initPassword}'`);
      const event: UploadInput = {
        type: 'uploadAll',
        url: environment.apiUrlOne + '/upload/file',
        method: 'POST',
        withCredentials: true,
        data: {
          pass: this.encPassword,
          initPass: this.initPassword,
        }
      };
      this.uploadInput.emit(event);
    } else if (output.type === 'addedToQueue' && typeof output.file !== 'undefined') { // add file to array when added
      this.files.push(output.file);
    } else if (output.type === 'uploading' && typeof output.file !== 'undefined') {
      // update current data in files array for uploading file
      const index = this.files.findIndex(file => typeof output.file !== 'undefined' && file.id === output.file.id);
      this.files[index] = output.file;
    } else if (output.type === 'removed') {
      // remove file from array when removed
      this.files = this.files.filter((file: UploadFile) => file !== output.file);
    } else if (output.type === 'removedAll') {
      this.files = [];
      this.uploadToken = '';
      this.uploadFileName = '';
    } else if (output.type === 'dragOver') {
      this.dragOver = true;
    } else if (output.type === 'dragOut') {
      this.dragOver = false;
    } else if (output.type === 'drop') {
      this.dragOver = false;
    } else if (output.type === 'cancelled') {
      this.files = [];
    } else if (output.type === 'done') {
      this.state.setProgress(false);
      const response: any = output.file.response;
      if (output.file.responseStatus === 201) {
        // done!
        this.files = [];
        this.uploadToken = response.id;
        this.uploadFileName = output.file.name;
        console.log(`token: '${this.uploadToken}'; fileName: '${this.uploadFileName}'`);
        new MessageUtils().showSuccess(this.snackBar, response.message);
      } else {
        console.log(response);
        new MessageUtils().showError(this.snackBar, response.detail);
      }

    }
  }

  // public searchForTags = (text: string): Observable<AutoCompleteModel[]> => {
  //   return this.service.searchTags(text).pipe(map(a => {
  //     // change the type of the array to meet the 'expectations' of ngx-chips
  //     return this.mapAutocomplete(a.result, TagType.Tag);
  //   }));
  // }

  // public searchForSenders = (text: string): Observable<AutoCompleteModel[]> => {
  //   return this.service.searchSenders(text).pipe(map(a => {
  //     // change the type of the array to meet the 'expectations' of ngx-chips
  //     return this.mapAutocomplete(a.result, TagType.Sender);
  //   }));
  // }

  public onClearUploadedFile() {
    this.uploadInput.emit({ type: 'removeAll' });
  }

  public isFormValid() {
    if (this.documentTitle !== '' && this.uploadFileName !== '' && this.uploadToken !== ''
      && this.selectedSenders.length > 0) {
      return true;
    }
    return false;
  }


  // -------------------------------------------------------------------------------------------







  // -----------------------------------------------------------------------------------------


  addSender(event: MatChipInputEvent): void {
    this._add(event, this.selectedSenders, this.senderCtrl);
  }

  selectedSender(event: MatAutocompleteSelectedEvent): void {
    this._selected(event, this.selectedSenders, this.senderCtrl, this.tagInput);
  }

  removeSender(input: string): void {
    this._remove(input, this.selectedSenders);
  }


  addTag(event: MatChipInputEvent): void {
    this._add(event, this.selectedTags, this.tagCtrl);
  }

  selectedTag(event: MatAutocompleteSelectedEvent): void {
    this._selected(event, this.selectedTags, this.tagCtrl, this.tagInput);
  }

  removeTag(input: string): void {
    this._remove(input, this.selectedTags);
  }




  private _add(event: MatChipInputEvent, target: string[], form: FormControl): void {
    const input = event.input;
    const value = event.value;

    // Add our value
    if ((value || '').trim()) {
      target.push(value.trim());
    }

    // Reset the input value
    if (input) {
      input.value = '';
    }

    form.setValue(null);
  }

  private _selected(event: MatAutocompleteSelectedEvent, target: string[], form: FormControl, elem: ElementRef<HTMLInputElement>): void {
    target.push(event.option.viewValue);
    elem.nativeElement.value = '';
    form.setValue(null);
  }

  private _remove(input: string, target: string[]): void {
    const index = target.indexOf(input);
    if (index >= 0) {
      target.splice(index, 1);
    }
  }

  // private mapAutocomplete(items: string[], type: TagType): AutoCompleteModel[] {
  //   const autocompletion: AutoCompleteModel[] = [];
  //   items.forEach(x => {
  //     const item = new AutoCompleteModel();
  //     item.display = x;
  //     item.value = x;
  //     item.type = type;
  //     autocompletion.push(item);
  //   });
  //   return autocompletion;
  // }

  // private convertSenderAndTags() {
  //   if (this.selectedSenders) {
  //     this.senders = [];
  //     this.selectedSenders.forEach(item => {
  //       this.senders.push(item.display);
  //     });
  //   }

  //   if (this.selectedTags) {
  //     this.tags = [];
  //     this.selectedTags.forEach(item => {
  //       this.tags.push(item.display);
  //     });
  //   }
  // }
}
