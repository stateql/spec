User:

- id is text

- name is text

- friends is many User thru befriendedBy

- befriendedBy is many User thru friends

- popularity is number thru count(friends)

- authoredTasks is many Task thru author

- assignedTasks is many Task thru assignees

- countTasksCompleted is number thru count(assignedTasks.completionStatus == on)

Task:

- id is text

- title is text

- content is text

- attachments is many file

- completionStatus is switch

- author is User thru authoredTasks

- assignees is many User thru assignedTasks

- summary is text thru summarize(content length:”short”)

- addFile is action thru put(attachments)

- removeFile is action thru drop(attachments :key)

- markComplete is action thru set(completionStatus value:on)

- parentTasks is many Task and parents

- subtasks is many Task and children

- subtask is action thru put(subtasks)

- relatedTasks is many Task and self

- priority is text thru either("high" "medium" "low")

- dueDate is date

- timeRemaining is seconds thru ticktock(duedate) // can be positive and negative

- startTime is timestamp

- endTime is timestamp

- timeSpent is seconds thru sum(subtasks.timeSpent ticktock(startTime endTime))
