package main

import "strings"

//Stop: []string{"tanaka:", "enquirer:", "reflector:", "prioritiser:", "planner:", "lister:", "decider:", "policy-decider:", "criticiser:", "recaller:", "tokensniffer:", "host:"},

var enquirer = ""

// the key for the map is the conversation history thread for that agent (int)
var previousconversations map[int]string = make(map[int]string)

var instruction = `You are an policy decider agent, you use the conversation history to decide which agent to run next. Here are the rules for the policy-decider agent:`
var rule0 = `If the last messenger in the thread is "tanaka:" and it also has a step number of 0 then the output of the policy-decider is the "summariser:" agent which, respond with the number 1, here is an example answer "policy-decider: 1".`
var rule1 = `if the last messenger is the thread is "summariser:" then the output of the policy-decider is the "criticiser:" agent which the index key is 2. Respond with the number 2, here is an example answer "policy-decider: 2".`
var rule2 = `if the last messenger in the thread is "criticiser:" then the output of the policy-decider is the "enquirer:" agent which the index key is 3. Respond with the number 3, here is an example answer "policy-decider: 3".`
var rule3 = `if the last messenger in the thread is "enquirer:" then the output of the policy-decider is the "prioritiser:" agent which the index key is 4. Respond with the number 4, here is an example answer "policy-decider: 4".`
var rule4 = `if the last messenger in the thread is "prioritiser:" then the output of the policy-decider is the "planner:" agent which the index key is 5. Respond with the number 5, here is an example answer "policy-decider: 5".`
var rule5 = `if the last messenger in the thread is "planner:" then the output of the policy-decider is the "lister:" agent which the index key is 6. Respond with the number 6, here is an example answer "policy-decider: 6".`
var rule6 = `if the last messenger in the thread is "lister:" then the output of the policy-decider is the "decider:" agent which the index key is 7. Respond with the number 7, here is an example answer "policy-decider: 7".`
var ruele8 = `if the last messenger in the thread is "decider:" then the output of the policy-decider is the "host:" agent which the index key is 8. Respond with the number 8, here is an example answer "policy-decider: 8".`
var rule9 = `if the last messenger in the thread is "host:" then the output of the policy-decider is 10 which ends the task and. Respond with the number 9, here is an example answer "policy-decider: 10".`

//every agent when ran runs with a recaller agent (except the agent that is ran at the step 1) to analyse the conversation history for the answer, the agent that handles the task and then the criticiser agent which is then recursively fed back to the thread of the agent that is handling the task to return the final answer of the task to the next agent.

var rules = []string{instruction, rule0, rule1, rule2, rule3, rule4, rule5, rule6}

// how can this slice be joined together to make a string that can be used as a prompt for the gpt3 api?
func joinrules() string {
	var rulesstring string
	for _, rule := range rules {
		rulesstring += rule
	}
	return rulesstring
}

func policyDeciderPrompt(conversationHistory string) string {
	return joinrules() + "\n" + "here is the conversation history:" + conversationHistory
}

var policies = []Policies{
	{agent: "tanaka:", policyString: "policy-decider: 1", prompt: ""},
	{agent: "summariser:", policyString: "policy-decider: 2", prompt: ""},
	{agent: "criticiser:", policyString: "policy-decider: 3", prompt: ""},
	{agent: "enquirer:", policyString: "policy-decider: 4", prompt: ""},
	{agent: "prioritiser:", policyString: "policy-decider: 5", prompt: ""},
	{agent: "planner:", policyString: "policy-decider: 6", prompt: ""},
	{agent: "lister:", policyString: "policy-decider: 7", prompt: ""},
	{agent: "decider:", policyString: "policy-decider: 8", prompt: ""},
	{agent: "host:", policyString: "policy-decider: 9", prompt: ""},
}

func runPolicies() string {
	var response string
	var task string = "Create a policy for doing research on the topic of the task."
Loop:
	for {
		for idx, i := range policies {
			switch {
			case idx == 0:
				continue
			case idx == 1:
				var x = i.prompt + `here is the task: ` + task
				response = getResponse(x)
			case idx == 8:
				break Loop
			default:
				response = getResponse(i.prompt + response)
			}
		}
	}

	return response
}

type Policies struct {
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