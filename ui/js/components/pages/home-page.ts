import {
  LitElement,
  html,
  css,
  customElement,
  TemplateResult,
  CSSResult,
} from "lit-element";
import { UserService } from "../../service/User";
import "./../card-element";
import "./../register-form";

@customElement("home-page")
export class HomePage extends LitElement {
  username: string = "";
  email: string = "";
  userService: UserService = new UserService();

  static get styles(): CSSResult {
    return css`
      login-container {
        display: grid;
        justify-content: center;
      }
    `;
  }

  firstUpdated(): void {
    this.handleGetUserProfile();
  }

  handleGetUserProfile(): void {
    this.userService.getUser().subscribe({
      next: (result: any) => {
        if ((result as { error: boolean; message: any }).error) {
          return console.error(
            (result as { error: boolean; message: any }).message
          );
        }
        const { username, email } = result as UserService.UserProfile;
        this.username = username;
        this.email = email;
        this.requestUpdate();
      },
    });
  }

  displayHomePage(): TemplateResult {
    if (this.username) {
      return html` <body-element /> `;
    } else {
      return html`
        <card-element>
          <login-container> 
            <register-form />
          </login-container>
        </card-ement>
        `;
    }
  }

  render(): TemplateResult {
    return html` ${this.displayHomePage()} `;
  }
}
