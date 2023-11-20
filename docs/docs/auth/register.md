import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Registering Users

This document provides instructions on registering users using either the SDK or CLI. 
The register endpoint is unauthenticated and thus allows you to create users in Rig if registration is enabled in your project.

<hr class="solid" />

## Register

To register users in Rig, you can utilize the Register endpoint. When registering a user, you must supply a password and an identifier. The identifiers can be usernames, emails, or phone numbers.

To control the enabled identifiers in your system, please refer to the [login methods section](/auth/auth-settings).

Make sure to set the `ProjectId` field to match your project ID.

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.Authentication().Register(ctx, connect.NewRequest(&authentication.RegisterRequest{
    Method: &authentication.RegisterRequest_UserPassword{
        UserPassword: &authentication.UserPassword{
            Password:   "YourPassword1234!",
            Identifier: &model.UserIdentifier{Identifier: &model.UserIdentifier_Email{Email: "johndoe@acme.com"}},
            ProjectId:     "YOU-PROJECT-ID",
        },
    },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully registered user with id %s \n", resp.Msg.GetUserId())
log.Printf("generated token pair: \nAccess token: %s\nRefresh token: %s \n", resp.Msg.GetToken().GetAccessToken(), resp.Msg.GetToken().GetRefreshToken())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.auth.register({
  method: {
    case: "userPassword",
    value: {
      identifier: {
        identifier: {
          case: "email",
          value: "johndoe@acme.com",
        },
      },
      password: "YourPassword1234!",
      projectId: "YourProjectId",
    },
  },
});
console.log(`successfully registered user with id ${resp.userId}`);
console.log(
  `generated token pair: \nAccess token: %s\nRefresh token: ${resp.token?.accessToken}`,
);
```

</TabItem>
</Tabs>

When a user is created using the username/password method, the response will include an access token and a refresh token pair. However, if a user is created using the email/password or phone/password method, that user must [verify their account](/auth/login#verify-email) before they can log in.
