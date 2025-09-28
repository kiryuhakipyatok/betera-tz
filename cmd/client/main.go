package main

import "betera-tz/client"

func main() {
	client := client.CreateClient()
	id := client.CreateNewTask("testTitle2", "testDescription2")
	if id != nil {
		client.UpdateTaskStatus(*id, "done")
		client.GetTaskById(*id)
	}
	page := 1
	amount := 10
	statusFilter := "done"
	client.FetchTasks(&amount, &page, &statusFilter)
}
