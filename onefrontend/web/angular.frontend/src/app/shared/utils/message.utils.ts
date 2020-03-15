import { MatSnackBar } from '@angular/material/snack-bar';

export class MessageUtils {

    public showError(snackBar: MatSnackBar, error: any) {
        console.error('Got error: ' + error);
        if (error.message) {
            this.showMessage(snackBar, 'Error: ' + error.message, 'error', -1, 'Dismiss!');
        } else {
            this.showMessage(snackBar, 'Error: ' + error, 'error', -1, 'Dismiss!');
        }
    }

    public showSuccess(snackBar: MatSnackBar, message) {
        this.showMessage(snackBar, message, 'success', 1500, '');
    }

    public showMessage(snackBar: MatSnackBar, message, type: string, duration: number, closeMessage: string) {
        if (duration > 0) {
            const snackBarRef = snackBar.open(message, closeMessage,
                {
                    duration: duration,
                    panelClass: [type]
                });
        } else {
            const snackBarRef = snackBar.open(message, closeMessage,
                {
                  panelClass: [type]
                });
        }
    }
}
