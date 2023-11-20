import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Implementing User Login
This document guides how to log in users using the SDK or CLI and perform email or phone number verification. You will learn the necessary steps to authenticate users and ensure the validity of their email or phone numbers.

Please note that this guide does not include instructions on how to authenticate users using OAuth providers such as "Login with Google". If you are interested in implementing authentication with social providers, please refer to the [social login section](/auth/social-login) for detailed instructions.

<hr class="solid" />

## Login

To authenticate users, you can utilize the `Login` endpoint. When logging in a user, you must provide their password along with an identifier. The supported identifiers include email addresses, phone numbers, and usernames.

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.Authentication().Login(ctx, connect.NewRequest(&authentication.LoginRequest{
    Method: &authentication.LoginRequest_UserPassword{
        UserPassword: &authentication.UserPassword{
            Password:   "Test1234!",
            Identifier: &model.UserIdentifier{Identifier: &model.UserIdentifier_Email{Email: "johndoe@acme.com"}},
            ProjectId:     "MyProjectId",
        },
    },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("generated token pair: \nAccess token: %s\nRefresh token: %s \n", resp.Msg.Token.AccessToken, resp.Msg.Token.RefreshToken)
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.auth.login({
  method: {
    case: "userPassword",
    value: {
      identifier: {
        identifier: {
          case: "email",
          value: "johndoe@acme.com",
        },
      },
      password: "Test1234!",
      projectId: "MyProjectId1234",
    },
  },
});
console.log("generated token pair", resp.token);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig auth login --user --password
```

Example:

```sh
rig auth login -u johndoe@acme.com -p Test1234!
```

</TabItem>
</Tabs>

If a user attempts to log in using the email/password or phone/password method and **account verification is enabled**, that user must [verify their account](/auth/login#account-verification) before the system generates an access/refresh token.

If a user logs in without verifying their identifier, Rig will send an email or text message to the user's contact information (used for verification) and return an error in response.

<hr class="solid" />

## Account Verification

### Verify Email

When a user creates an account using the email/password method and account verification is enabled, we send them an email containing instructions for verification. This email includes a verification code and provides information on how to verify their account. If you wish to manage email templates, please refer to the [verification templates section](/auth/auth-templates).

To verify a user's email, you can utilize the `VerifyEmail` endpoint in your backend:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.Authentication().VerifyEmail(ctx, connect.NewRequest(&authentication.VerifyEmailRequest{
    Code:   "CODE-FROM-EMAIL",
    Email:  "johndoe@acme.com",
    ProjectId: "MyProjectId234",
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully verified email")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
if _, err := client.Authentication().VerifyEmail(ctx, connect.NewRequest(&authentication.VerifyEmailRequest{
    Code:   "CODE-FROM-EMAIL",
    Email:  "johndoe@acme.com",
    ApiKey: "MyProjectId1234",
}))
console.log("successfully verified email")
```

</TabItem>
</Tabs>

Once a user has successfully verified their email, they are now able to log in to their account.

### Verify Phone Number

Similar to email authentication, when logging in with a phone number and password, a verification step may be triggered. To handle text message verification templates, please refer to the [verification templates section](/auth/auth-templates) for management options.

To verify a user's phone number, you can utilize the `VerifyPhoneNumber` endpoint in your backend:
<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.Authentication().VerifyPhoneNumber(ctx, connect.NewRequest(&authentication.VerifyPhoneNumberRequest{
    Code:   "CODE-FROM-EMAIL",
    PhoneNumber:  "+4522122798",
    ProjectId: "MyProjectId1234",
})); err != nil {
    log.Fatal(err)
}
log.Printf("successfully verified phone number")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.auth.verifyPhoneNumber({
    code:   "CODE-FROM-EMAIL",
    phoneNumber:  "+4522122798",
    projectId: "MyProjectId1234",
})
console.log("successfully verified phone number")
```

</TabItem>
</Tabs>
Once a user has successfully verified their phone number, they are now able to log in to their account.
