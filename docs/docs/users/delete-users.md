---
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Deleting Users

This document provides instructions on how to delete users using the SDK or CLI.

<hr class="solid" />

## Deleting a Single User

To delete a single user from Rig, you can utilize the `Delete` endpoint:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
userID := "" // NOTE: insert a specifc userID here
if _, err := client.User().Delete(ctx, connect.NewRequest(&user.DeleteRequest{
  UserId: userID,
})); err != nil {
  log.Fatal(err)
}
log.Println("successfully deleted user")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
userID = ""; // NOTE: insert a specifc userID here
await client.user.delete({
  userId: userID,
});
console.log("successfully deleted user");
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user delete [user-id | {email|username|phone}]
```

Example:

```sh
rig user delete john@acme.io
```

</TabItem>
</Tabs>
