package main

func interpreter(prompt string) string {
	var a string = `human: Does the following prompt include anything to do with any of these topics? Prompt: "` + prompt + `" The choices are: 
1. Food making
2. Coding
3. Music

one shot examples of prompts and their responses:
human: can you suggest something to eat? 
ai: yes, food.
human: can you help plan our dinner?
ai: yes, food.

if no then write no, if you don't know just write nil 

ai: 
\n
`

	return a
}

/*
	human: what would be the best answer the question with the goal of assisting with the task? // this could help better answer the question if extra information is not needed like an example of a plugin.

	ai:

	human: what other context would you like to know to better answer the question.

	ai:
*/
