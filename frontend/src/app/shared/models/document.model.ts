import * as moment from 'moment';

export class MyDmsDocument {
  id: string;
  title: string;
  fileName: string;
  encodedFilename: string;
  alternativeId: string;
  previewLink: string;
  uploadFileToken: string;
  amount: number;
  created: Date;
  modified: Date;
  tags: Array<string>;
  senders: Array<string>;
  dateHuman: string;
  invoiceNumber: string;
  update: string;

  public get dateFormat(): string {
    const dateFormat = moment(this.created).format('DD.MM.YYYY');
    return dateFormat;
  }

  public get lastDate(): Date {
    let date1: Date = moment(this.modified).toDate();
    if (!date1) {
      date1 = new Date(1970, 1, 1);
    }
    let date2: Date = moment(this.created).toDate();
    if (!date2) {
      date2 = new Date(1970, 1, 1);
    }

    let date: Date = null;

    if (date1 > date2) {
      date = date1;
    } else {
      date = date2;
    }
    return date;
  }

  /**
   * Number.prototype.format(n, x, s, c)
   *
   * @param integer n: length of decimal
   * @param integer x: length of whole part
   * @param mixed   s: sections delimiter
   * @param mixed   c: decimal delimiter
   */
  public formatAmount(n: number, x: number, s: string, c: string): string {
    const re = '\\d(?=(\\d{' + (x || 3) + '})+' + (n > 0 ? '\\D' : '$') + ')',
    num = this.amount.toFixed(Math.max(0, n));
    return (c ? num.replace('.', c) : num).replace(new RegExp(re, 'g'), '$&' + (s || ','));
  }
}
