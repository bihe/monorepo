import { Component } from '@angular/core';
import { ApplicationState } from 'src/app/shared/service/application.state';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'app';
  currentRoute = '';
  navExpanded = false;

  constructor(private stateService: ApplicationState) {
    this.stateService.getRoute().subscribe(
      data => {
        this.currentRoute = data;
      }
    );
  }

  isCurrentRout(route: string): boolean {
    let isCurrent = route === this.currentRoute;
    if (!isCurrent) {
      isCurrent = this.currentRoute.startsWith(route);
    }
    return isCurrent;
  }

  toggleNavbar() {
    this.navExpanded = !this.navExpanded;
  }
}
