import InstallRig from '../../src/markdown/prerequisites/install-rig.md'
import SetupSdk from '../../src/markdown/prerequisites/setup-sdk.md'
import SetupCli from '../../src/markdown/prerequisites/setup-cli.md'
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Implementing User Logout

This document provides instructions on how to log out a user using the SDK. By logging out, you will invalidate the user's authentication session, including using their refresh token to create new token pairs, preventing their further usage.

<hr class="solid" />

## Logout a User

To log out users in your backend, you can utilize the `Logout` endpoint. When logging out a user, you will need to provide the access token associated with that user, which needs to be saved in the context.

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
ctx = context.WithValue(ctx, "Authorization Bearer", "USER-ACCESS-TOKEN")
if _, err := client.Authentication().Logout(ctx, connect.NewRequest(&authentication.LogoutRequest{})); err != nil {
    log.Fatal(err)
}
return nil
```

</TabItem>
</Tabs>
