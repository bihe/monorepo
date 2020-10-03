import { Pipe, PipeTransform } from '@angular/core';
import * as moment from 'moment';

@Pipe({
    name: 'dfmt'
})
export class DateFormatPipe implements PipeTransform {
  transform(val: any, args: any) {
    if (args === undefined) {
      return val;
    }

    try {
      if (val) {
        const date: string = val.toString();
        const format: string = args.toString();
        const dateFormat = moment.utc(date).local().format(format);
        return dateFormat;
      }
    } catch (EX) {
      console.error('Could not format date: ' + EX);
    }
    return val;
  }
}
