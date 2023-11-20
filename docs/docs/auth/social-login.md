

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Implementing Social Login

In this document, you’ll learn how to authenticate users using a social provider, through the SDK.

Rig integrates with multiple social providers such as `Google`, `Facebook`, and `GitHub`. You can use these providers to implement functionality like *Login with Google* or *Register with GitHub*. 

<hr class="solid" />

## Prerequisites
### Setup your social providers
It is assumed that you configured at least one social provider in your Rig backend and have configured one or more redirect addresses. If not, you can follow the guide on [how to manage your social providers](/auth/auth-settings).

<hr class="solid" />

## Implementation
### 1. Generate the Auth Config
Use the `AuthConfig` endpoint to fetch your project's public auth configuration. To generate auth links for your social providers, pass a `RedirectAddr` to tell your social provide (eg. Google) where to redirect your user after logging in:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
authConfigResp, err := client.Authentication().GetAuthConfig(ctx, connect.NewRequest(&authentication.GetAuthConfigRequest{
  RedirectAddr: "YOUR-REDIRECT-ADDR",
  ProjectId:    "YOUR-PROJECT-ID",
}))
if err != nil {
  log.Fatal(err)
}
log.Println(authConfigResp.Msg.GetOauthProviders())
```
</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig auth get-auth-config {project-id | project-name} --redirect-addr
```

Example:
```sh
rig auth get-auth-config acme-project -r localhost:3000/callback
```

</TabItem>
</Tabs>

The response returned contains a list of `OauthProviders`. Each provider has a property called `ProviderUrl` which is a URL generated for each provider that takes the user to the provider's auth page.

### 2. Redirect the User
Redirect your user to the intended `ProviderUrl`. You can do this from either your backend or frontend. An example in Golang is provided below:
<Tabs>
<TabItem value="go" label="Golang SDK">

```go
authConfigResp, err := client.Authentication().GetAuthConfig(ctx, connect.NewRequest(&authentication.GetAuthConfigRequest{
  RedirectAddr: "YOUR-REDIRECT-ADDR",
  ProjectId:    "YOUR-PROJECT-ID",
}))
if err != nil {
  log.Fatal(err)
}
// highlight-next-line
http.Redirect(w, r, authConfigResp.Msg.OauthProviders[0].GetProviderUrl(), http.StatusPermanentRedirect)
```
</TabItem>
</Tabs>

### 3. Manage the response
When the user has successfully logged in, he will be redirected to the provided `RedirectAddr`. The will look something like this: `http://localhost:3000/callback?access_token=access-token-value&refresh_token=refresh-token-value` depending on your configured redirect URL.

From the URL, you can fetch the `access_token` and `refresh_token` and use those to authorize your user. 