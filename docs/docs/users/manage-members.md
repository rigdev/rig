---
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Members using the SDK or CLI

This document provides instructions on managing members in groups using the SDK or CLI. It covers various operations such as adding and removing members, as well as fetching members from specific groups. By following the guidelines provided, you will learn how to effectively manage group members using the SDK functionalities.

<hr class="solid" />

## Adding a Member

To add members to groups in your backend, you can utilize the `AddMember` endpoint. When making a request to add one or more members, include the `UUIDs` of the users and the group in your request. By sending this information along with the request, you can successfully add users as members of a specific group in your backend:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := ""   // NOTE: insert a specifc groupID here
userIDOne := "" // NOTE: insert a specifc userID here
userIDTwo := "" // NOTE: insert a specifc userID here
if _, err := client.Group().AddMember(ctx, connect.NewRequest(&group.AddMemberRequest{
    GroupId: groupID,
    UserIds:  []string{userIDOne, userIDTwo},
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully added users to group")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const groupId = ""; // NOTE: insert a specifc groupID here
const userIdOne = ""; // NOTE: insert a specifc userID here
const userIdTwo = ""; // NOTE: insert a specifc userID here
await client.groupsClient.AddMember({
  groupId: groupId,
  userIds: [userIdOne, userIdTwo],
});
console.log("Successfully added users to a group");
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group add-member [user-id | {email|username|phone}]
```

Example:

```sh
rig group add-member john@acme.com
```

In the CLI it is only possible to add one member at a time. The available groups are then listed for selection.

</TabItem>
</Tabs>

<hr class="solid" />

## Removing a Member

To remove members from groups in your backend, you can make use of the `RemoveMember` endpoint. By sending the `UUID` of the user and group in your request, you can successfully remove a user as a member from the specified group in your backend:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := ""   // NOTE: insert a specifc groupID here
userIDOne := "" // NOTE: insert a specifc userID here
if _, err := client.Group().RemoveMember(ctx, connect.NewRequest(&group.RemoveMemberRequest{
    GroupId: groupID,
    UserId:  userIDOne,
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully removed user from group")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
groupID = ""; // NOTE: insert a specifc groupID here
userIDOne = ""; // NOTE: insert a specifc userID here
await client.groupsClient.removeMember({
  groupId: groupID,
  userId: userID,
});
console.log("successfully removed user from group");
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group add-member [user-id | {email|username|phone}]
```

Example:

```sh
rig group remove-member john@acme.com
```

Groups that the user is a member of are then listed for selection.
</TabItem>
</Tabs>

<hr class="solid" />

## Listing Members in a Group

To retrieve a list of members in groups from your backend, you can utilize the `ListMembers` endpoint:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
groupID := "" // NOTE: insert a specifc groupID here
resp, _ := client.Group().ListMembers(ctx, connect.NewRequest(&group.ListMembersRequest{
    GroupId: groupID,
    Pagination: &model.Pagination{
        Offset: 10,
        Limit:  10,
    },
}))
log.Printf("successfully fetched %d members. Total amount is: %d", len(resp.Msg.GetMembers()), resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
groupID = ""; // NOTE: insert a specifc groupID here
const resp = await client.groups.listMembers({
  groupId: groupID,
  pagination: {
    offset: 10,
    limit: 10,
  },
});
console.log(
  `successfully fetched ${resp.members.length} members. Total amount is: ${resp.total}`,
);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig group list-members [group-id | group-name] --offset --limit --json
```

Example:

```sh
rig group list-members editors -o 10 -l 10 --json
```

</TabItem>
</Tabs>

The above query will fetch `10` members in your group, starting from member number `10`. The total amount of members is returned as well.
