package text

import "gopkg.in/jdkato/prose.v2"

// GetSentencesFromString returns each sentence in an input string, or an empty slice if there are no sentences in the input
func GetSentencesFromString(rawInput string) ([]string, error) {
	doc, err := prose.NewDocument(rawInput, prose.WithExtraction(false), prose.WithTagging(false), prose.WithTokenization(false))
	if err != nil {
		return nil, err
	}

	sentenceTexts := []string{}
	sentences := doc.Sentences()
	for _, sentence := range sentences {
		sentenceTexts = append(sentenceTexts, sentence.Text)
	}

	return sentenceTexts, nil
}