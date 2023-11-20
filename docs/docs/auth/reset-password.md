

import InstallRig from '../../src/markdown/prerequisites/install-rig.md'
import SetupSdk from '../../src/markdown/prerequisites/setup-sdk.md'
import SetupCli from '../../src/markdown/prerequisites/setup-cli.md'
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Resetting Passwords
This document provides instructions on how to reset a user's password using the SDK or CLI. The user will receive a reset password link either through SMS or email. Please note that for security reasons, we require the user to be verified before they can perform this action.

If a user has not yet verified their account and is unable to remember their password, please utilize the Dashboard to update the user's password.

<hr class="solid" />

## Implementation
### 1. Send Reset Password Link
To send a reset password template, you can utilize the `SendPasswordReset` endpoint. Make the API call with your `Project ID` and the user's `Identifier`. The available identifiers include usernames, emails, and phone numbers. Make sure to include one of these fields when calling the client:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.Authentication().SendPasswordReset(ctx, connect.NewRequest(&authentication.SendPasswordResetRequest{
    Identifier: &model.UserIdentifier{
        Identifier: &model.UserIdentifier_Email{Email: "johndoe@acme.com"},
    },
    ProjectId: "YOUR-PROJECT-ID",
})); err != nil {
    log.Fatal(err)
}
```
</TabItem>
</Tabs>

By performing this action, the reset password flow will be triggered in the backend, and the user will receive instructions via email or text message. You can customize the reset password templates by accessing the [auth templates section](/auth/auth-templates).

### 2. Reset Password
To reset the password, you can utilize the `ResetPassword` endpoint. Make the API call with your `Project ID`, the user's `Identifier`, the `NewPassword`, and the `Code` that was sent to the user in the previous step. The available identifiers include usernames, emails, and phone numbers. Make sure to include these fields when making the call to the client:
<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.Authentication().ResetPassword(ctx, connect.NewRequest(&authentication.ResetPasswordRequest{
    Code:        "CODE-FROM-EMAIL-OR-SMS",
    NewPassword: "YOUR-NEW-PASSWORD-1234!",
    Identifier: &model.UserIdentifier{
        Identifier: &model.UserIdentifier_Email{Email: "johndoe@acme.com"},
    },
    ProjectID: "YOUR-PROJECT-ID",
})); err != nil {
    log.Fatal(err)
}
```
</TabItem>
</Tabs>

The user can not login with their new password.