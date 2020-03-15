import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogConfig, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';

@Component({
  selector: 'app-delete-confirmation',
  templateUrl: './confirmation.component.html',
  styleUrls: ['./confirmation.component.css']
})
export class ConfirmDeleteDialogComponent implements OnInit {

    name = '';

    constructor(
        private dialog: MatDialogRef<ConfirmDeleteDialogComponent>,
        @Inject(MAT_DIALOG_DATA) private dialogData: any) {
        this.name = dialogData.name;
    }

    public static getDialogConfig(name: string): MatDialogConfig {
        const dialogConfig: MatDialogConfig = {
            disableClose: false,
            width: '440px',
            height: '240px',
            data: {
                name: name
            }
        };
        return dialogConfig;
    }

    public onConfirmDelete() {
        this.dialog.close(true);
    }

    ngOnInit() {
    }
}
