---
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Groups using the SDK or CLI

This document provides instructions on managing groups using the SDK or CLI. It covers various operations such as fetching and updating groups, as well as fetching groups for a specific user.

<hr class="solid" />

## Fetching Groups

### Getting a Group by UUID

To retrieve groups in Rig, you can utilize the `Get` endpoint by including the unique `UUID` of the group in your request:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := "" // NOTE: insert a specifc groupID here
resp, err := client.Group().Get(ctx, connect.NewRequest(&group.GetRequest{
    GroupId: groupID,
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully fetched group: %s \n", resp.Msg.GetGroup().GetName())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const groupID = ""; // NOTE: insert a specifc groupID here
const resp = await client.groupsClient.get({
  groupId: groupID,
});
console.log(`successfully fetched group: ${resp.group}`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group get [group-id | group-name] --json
```

Example:

```sh
rig group get admins
```

</TabItem>
</Tabs>

### Listing Groups

To list groups in Rig, you can utilize the `List` endpoint. Pagination can be implemented using the `Offset` and `Limit` fields. The `Offset` field determines the starting point of the list, while the `Limit` field specifies the maximum number of groups to retrieve in each request.
<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, _ := client.Group().List(ctx, connect.NewRequest(&group.ListRequest{
    Pagination: &model.Pagination{
        Offset: 10,
        Limit:  10,
    },
}))
log.Printf("successfully fetched %d groups. Total amount is: %d \n", len(resp.Msg.GetGroups()), resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.groupsClient.list({
  pagination: {
    offset: 10,
    limit: 10,
  },
});
console.log(
  `successfully fetched ${resp.groups.length} groups. Total amount is: ${resp.total}`,
);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig group list --offset --limit --json
```

Example:

```sh
rig group list -o 10 -l 10 --json
```

</TabItem>
</Tabs>

The above query will fetch `10` groups in your project, starting from group number `10`. The total amount of groups is returned as well.

### Listing Groups for a Specific User

To retrieve a list of groups associated with a specific user in your backend, you can utilize the `ListGroupsForUser` endpoint:
<Tabs>
<TabItem value="go" label="Golang SDK">

```go
userID := "" // NOTE: insert a specifc userID here
resp, _ := client.Group().ListGroupsForUser(ctx, connect.NewRequest(&group.ListGroupsForUserRequest{
    UserId: userID,
    Pagination: &model.Pagination{
        Offset: 10,
        Limit:  10,
    },
}))
log.Printf("successfully fetched %d groups. Total amount is: %d \n", len(resp.Msg.GetGroups()), resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const userID = ""; // NOTE: insert a specifc userID here
const resp = await client.groupsClient.listGroupsForUser({
  userId: userID,
  pagination: {
    offset: 10,
    limit: 10,
  },
});
console.log(
  `successfully fetched ${resp.groups.length} groups. Total amount is: ${resp.total}`,
);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig group list-groups-for-user [user-id | {email|username|phone}] --offset --limit --json
```

Example:

```sh
rig group list-groups-for-user john@acme.com -o 10 -l 10
```

</TabItem>
</Tabs>

The above query will fetch `10` groups for the specific user, starting from group number `10`. The total amount of groups for the user is returned as well.

## Updating Groups

To update group information, you can utilize the `Update` endpoint. It is important to note that you can include one or multiple `Updates` in your request to modify multiple fields simultaneously. Here's an example of how it can be done:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := "" // NOTE: insert a specifc groupID here
if _, err := client.Group().Update(ctx, connect.NewRequest(&group.UpdateRequest{
    GroupId: groupID,
    Updates: []*group.Update{
        {   // To update name
            Field: &group.Update_Name{Name: "editors"},
        },
        {   // To insert/update metadata key-value pair
            Field: &group.Update_SetMetadata{SetMetadata: &model.Metadata{Key: "role", Value: []byte("1234")}},
        },
    },
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully updated group")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.groupsClient.update({
  groupId: groupID,
  updates: [
    {
      // To update name
      field: {
        case: "name",
        value: "editors",
      },
    },
    {
      field: {
        case: "setMetadata",
        value: {
          key: "role",
          value: new TextEncoder().encode("1234"),
        },
      },
    },
  ],
});
console.log("successfully updated group");
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group update [group-id | group-name] --field --value
```

Example:

```sh
rig group update admins -f name -v editors
rig group update admins -f set-meta-data -v '{"key":"role","value":"1234"}'
```

</TabItem>
</Tabs>
