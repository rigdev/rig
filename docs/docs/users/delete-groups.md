---
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Deleting Groups

This document provides instructions on how to delete groups using the SDK or CLI.

<hr class="solid" />

## Delete a Group

You can delete a group on your backend using the `Delete` endpoint:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := "" // NOTE: insert a specifc groupID here
if _, err := client.Group().Delete(ctx, connect.NewRequest(&group.DeleteRequest{
    GroupId: groupID,
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully deleted group")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const groupID = ""; // NOTE: insert a specifc groupID here
await client.groupsClient.delete({
  groupId: groupID,
});
console.log("successfully deleted group");
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig group delete [group-id | group-name]
```

Example:

```sh
rig group delete admins
```

</TabItem>
</Tabs>
