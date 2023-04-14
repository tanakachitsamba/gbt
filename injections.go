package main

import "strings"

//every agent when ran runs with a recaller agent (except the agent that is ran at the step 1) to analyse the conversation history for the answer, the agent that handles the task and then the criticiser agent which is then recursively fed back to the thread of the agent that is handling the task to return the final answer of the task to the next agent.

// this is the general task agent policy, others will need to be created.
var policy = []Policy{
	{agent: "summariser:", prompt: ``},
	{agent: "criticiser:", prompt: ``},
	{agent: "enquirer:", prompt: ``},
	{agent: "prioritiser:", prompt: ``},
	{agent: "planner:", prompt: ``},
	{agent: "lister:", prompt: ``},
	{agent: "decider:", prompt: ``},
	{agent: "executor:", prompt: ``},
}

func runPolicies(instruction string, ifBlockedTask bool) string {
	var response string
	//var nexttask string = `create a report on how this agent could useful.`

	//var waitingForResponse bool
	// todo: this needs to be a db for backing up the conversation history
	var conversationThreads []ConversationThread = make([]ConversationThread, 0)
Loop:
	for {
		var (
			conversationThread     ConversationThread
			summary                string
			qoutes                 string
			ifPreviousConversation bool
		)
		if ifBlockedTask {
			// this needs to be batched to multiple goroutines to speed up the process
			for _, x := range conversationThreads {
				var input string = `Does this conversation have an answer to this following query: "` + instruction + `". Answer with a "Yes" or "No" or If you don't know just say I don't know. Here is the conversation: ` + x.conversation
				response = getResponse(input)
				if filterString(response, "yes") {
					ifPreviousConversation = true
					summary = getResponse(`Read the conversation below and Summarise all the relevant information that relates to answering this query: "` + instruction + `". Here is the conversation: ` + x.conversation)
					qoutes = getResponse(`Read the conversation below and note all qoutes verbatim that relate to answering this query: ` + instruction + `". Here is the conversation: ` + x.conversation)
					break
				}
			}
		}
		for idx, i := range policy {
			switch {
			case idx == 0:
				var previousConversation = func(ifPreviousConversation bool) string {
					if !ifPreviousConversation {
						return ``
					}
					return `Here is the summaries and qoutes of the previous conversation history: ` + summary + qoutes
				}

				response = getResponse(i.prompt + instruction + previousConversation(ifPreviousConversation))
				conversationThread.conversation += i.prompt + instruction + previousConversation(ifPreviousConversation) + response
			case idx == 1:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 2:
				/*
					todo: enquier agent needs to be created to ask the user for the task
									response = getResponse(i.prompt + response)
					if filterString(response, "enquirer:") {
					}
				*/
				continue
			case idx == 3:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 4:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 5:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 6:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 7:
				response = getResponse(i.prompt + response)
				conversationThread.conversation += i.prompt + response
				conversationThreads = append(conversationThreads, conversationThread)
				break Loop
			}
		}
	}
	_ = conversationThreads

	return response
}

type ConversationThread struct {
	conversation string
}

type Policy struct {
	agent, policyString, prompt string
}

type String struct {
	key string
}

func filterString(input, query string) bool {
	return strings.Contains(input, query)
}

func expector(responseOutput string) {

	var prompt = `You are an agent that checks whether the output is as expected or not. ` + `the output should respond with an integer  or I dont know, if the ouput of the previous agent is not as expected then the current agent will return an error message. Here is an example answer "expectation error: should be an integer or I dont know but recieved something else". eitherwise return with just "passed". ` + "\n" + "here is the output to be reviewed:" + responseOutput

	res := getResponse(prompt)
	_ = res
}
