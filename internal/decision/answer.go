package decision

import (
	"fmt"
	"slices"
)

// Response is the body /v1/systemone answers with; only choice answers are read.
type Response struct {
	// Model is the model and version that answered, as the server names it.
	Model   string `json:"model"`
	Answers map[string]struct {
		Choice        string             `json:"choice"`
		Probabilities map[string]float64 `json:"probabilities"`
	} `json:"answers"`
}

// Option is one option of an answer and the probability the model gave it.
type Option struct {
	Name        string  `json:"name"`
	Probability float64 `json:"probability"`
}

// Answer is the model's choice for one question, with every option in the order asked.
type Answer struct {
	Choice  string   `json:"choice"`
	Options []Option `json:"options"`
}

// Verdict checks the answer to the verdict question against its options. An answer outside the options, or a
// missing one, is an error: a model that answers something else is not shown.
func Verdict(req Request, resp Response) (Answer, error) {
	a, ok := resp.Answers[QuestionVerdict]
	if !ok {
		return Answer{}, fmt.Errorf("no answer to %q", QuestionVerdict)
	}
	opts := req.Questions[QuestionVerdict].Criteria
	if !slices.ContainsFunc(opts, func(o option) bool { return o.Name == a.Choice }) {
		return Answer{}, fmt.Errorf("answer %q to %q is not an option", a.Choice, QuestionVerdict)
	}
	out := Answer{Choice: a.Choice}
	for _, o := range opts {
		p, ok := a.Probabilities[o.Name]
		if !ok {
			return Answer{}, fmt.Errorf("no probability for %q in %q", o.Name, QuestionVerdict)
		}
		out.Options = append(out.Options, Option{Name: o.Name, Probability: p})
	}
	return out, nil
}
