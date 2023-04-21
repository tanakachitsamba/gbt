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

func fetchExamples(prompt string) string {
	var examples string
	return prompt + examples
}

func handlingCritic(prompt string) string {
	// the first layer is to make sure that examples can be provided for icl fsl/msl
	var answer string = getResponse(fetchExamples(prompt), 0)
	var criticPrompt string = policy[1].prompt + answer

	var critic string = getResponse(fetchExamples(criticPrompt), 0.7)
	var lastPass string = getResponse(fetchExamples(criticPrompt+critic), 0)
	return lastPass
}

func computeProtocols() map[S]string {
	var protocols map[S]string = make(map[S]string, 0)

	protocols[S{key: "therapist"}] = `You are a therapist agent. 
Use evidence-based therapy techniques (e.g., CBT, DBT, mindfulness) to help users with therapist or couples therapist services or children's mental health. 
Pinpoint users' problems and deliver relevant guidance or advice.
Use the tone of the user from their sentiment for an accurate perception of user emotions. 
Foster a safe and supportive conversational atmosphere.
Identify users' issues and provide tailored advice or recommendations.
Accurately comprehend and empathize with users' emotions and feelings.
Provide emotional support and guidance to users in distress.`

	/*
	   protocols["chef"] = ``
	   protocols["teacher"] = ``
	   protocols["engineer"] = ``
	   protocols["programmer"] = ``
	   protocols["psychologist"] = ``
	   protocols["psychiatrist"] = ``
	   protocols["farmer"] = ``
	*/

	return protocols
}

type RunPolicies func(string, bool) string

// todo: certain protocols need to be saved so when used their memories can be loaded
func runPolicies(instruction string, loadMemory bool) string {
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

		// this looks for tasks that are blocked to see if the conversation history has an answer to the blocked task
		if loadMemory {
			// this needs to be batched to multiple goroutines to speed up the process
			for _, x := range conversationThreads {
				var input string = `Does this conversation have an answer to this following query: "` + instruction + `". Answer with a "Yes" or "No" or If you don't know just say I don't know. Here is the conversation: ` + x.conversation
				response = handlingCritic(input)
				if filterString(response, "Yes") {
					ifPreviousConversation = true

					var instruction1 string = `Read the conversation below and Summarise all the relevant information that relates to answering this query: "` + instruction + `". Here is the conversation: ` + x.conversation
					var instruction2 string = `Read the conversation below and note all qoutes verbatim that relate to answering this query: ` + instruction + `". Here is the conversation: ` + x.conversation
					summary = handlingCritic(instruction1)
					summary = "Here is the summaries of the previous conversation history: " + summary
					qoutes = handlingCritic(instruction2)
					qoutes = "Here is the qoutes of the previous conversation history: " + qoutes

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
					return summary + qoutes
				}

				response = handlingCritic(i.prompt + instruction + previousConversation(ifPreviousConversation))
				conversationThread.conversation += i.prompt + instruction + previousConversation(ifPreviousConversation) + response
			case idx == 1:
				response = handlingCritic(i.prompt + response)
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
				response = handlingCritic(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 4:
				response = handlingCritic(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 5:
				response = handlingCritic(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 6:
				response = handlingCritic(i.prompt + response)
				conversationThread.conversation += i.prompt + response
			case idx == 7:
				response = handlingCritic(i.prompt + response)
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

type S struct {
	key string
}

func filterString(input, query string) bool {
	return strings.Contains(input, query)
}

func expector(responseOutput string) {

	var prompt = `You are an agent that checks whether the output is as expected or not. ` + `the output should respond with an integer  or I dont know, if the ouput of the previous agent is not as expected then the current agent will return an error message. Here is an example answer "expectation error: should be an integer or I dont know but recieved something else". eitherwise return with just "passed". ` + "\n" + "here is the output to be reviewed:" + responseOutput

	res := getResponse(prompt, 0)
	_ = res
}

func countChars(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
