import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'ellipsis'
})
export class EllipsisPipe implements PipeTransform {
  transform(val: any, args: number) {
    if (args === undefined) {
      return val;
    }
    if (val === undefined) {
      return '';
    }
    if (!val || val === null) {
      return '';
    }

    if (val.length > args) {
      return val.substring(0, args) + '..';
    } else {
      return val;
    }
  }
}
