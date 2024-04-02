package examples

import "testing"

func TestExamples(t *testing.T) {
	ExampleSlogWithZlogHandler()
	ExampleZlogLogger()
	ExampleDevelopment()
	ExampleStacktrace()
	ExampleTimeFormatter()
	ExampleContextExtractorTraceContext()
	ExampleContextExtractorUserContext()
}
