package main

import (
	"encoding"
	"fmt"
	"sort"
	"strings"
)

type Chuncker struct{
	FileExt string

}


func(c *Chuncker) chunk(code string, tokenLimit int, encodingName string) map[int]string {

	chunks := make(map[int]string)

	currentChunk := ""

	tokenCount := 0
	lines := strings.Split(code, "\n")
	i := 0
	chunkNumber := 1
	startLine := 0

	content := []byte(code)
	codeParser := NewCodeParser(content, c.FileExt)
	_, err := codeParser.ParserAst()
	if err != nil{
		fmt.Print("code parse failed: ", err)
		return chunks
	}
	breakpoints := codeParser.GetLinesForPointsOfInterest(content)

	sort.Ints(breakpoints)

	for i < len(lines) {
		line := lines[i]
		newTokenCount := countTokens(line, encodingName)
		
		if tokenCount+newTokenCount > tokenLimit {
			stopLine := startLine
			for j := len(breakpoints) - 1; j >= 0; j--{
				if breakpoints[j] < i+1{
					stopLine = breakpoints[j]
					break
				}
			}
			if stopLine == startLine && !contains(breakpoints, i+1){
				tokenCount += newTokenCount
				i++
				continue
			}else if stopLine == startLine && contains(breakpoints, i+1){
				currentChunk = strings.Join(lines[startLine:i+1], "\n")
				if strings.TrimSpace(currentChunk) != ""{
					chunks[chunkNumber] = currentChunk
					chunkNumber++
				}
				tokenCount = 0 
				startLine = i+1
				i++
				continue
			}else{
				currentChunk = strings.Join(lines[startLine:stopLine], "\n")
				if strings.TrimSpace(currentChunk) != ""{
					chunks[chunkNumber] = currentChunk
					chunkNumber++
				}
				i = stopLine
				tokenCount = 0
				startLine = stopLine
				continue

			}
		}else{
			tokenCount += newTokenCount
			i++
		}
	
	}	
	
	// Append remaining code
	currentChunk = strings.Join(lines[startLine:], "\n")
	if strings.TrimSpace(currentChunk) != ""{
		chunks[chunkNumber] = currentChunk
	}

	return chunks
}


func countTokens(text, encodingName string) int{
	// todo finish me
	return len(strings.Fields(text))
}