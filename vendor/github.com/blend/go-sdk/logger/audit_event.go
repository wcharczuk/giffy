package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// NewAuditEvent returns a new audit event.
func NewAuditEvent(principal, verb, noun string) *AuditEvent {
	return &AuditEvent{
		ts:        time.Now().UTC(),
		flag:      Audit,
		principal: principal,
		verb:      verb,
		noun:      noun,
	}
}

// NewAuditEventListener returns a new audit event listener.
func NewAuditEventListener(listener func(me *AuditEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*AuditEvent); isTyped {
			listener(typed)
		}
	}
}

// AuditEvent is a common type of event detailing a business action by a subject.
type AuditEvent struct {
	heading   string
	ts        time.Time
	flag      Flag
	principal string
	verb      string
	noun      string
	subject   string
	property  string
	extra     map[string]string

	labels      map[string]string
	annotations map[string]string
}

// WithLabel sets a label on the event for later filtering.
func (ae *AuditEvent) WithLabel(key, value string) *AuditEvent {
	if ae.labels == nil {
		ae.labels = map[string]string{}
	}
	ae.labels[key] = value
	return ae
}

// Labels returns a labels collection.
func (ae *AuditEvent) Labels() map[string]string {
	return ae.labels
}

// WithAnnotation adds an annotation to the event.
func (ae *AuditEvent) WithAnnotation(key, value string) *AuditEvent {
	if ae.annotations == nil {
		ae.annotations = map[string]string{}
	}
	ae.annotations[key] = value
	return ae
}

// Annotations returns the annotations set.
func (ae *AuditEvent) Annotations() map[string]string {
	return ae.annotations
}

// WithFlag sets the audit event flag
func (ae *AuditEvent) WithFlag(f Flag) *AuditEvent {
	ae.flag = f
	return ae
}

// Flag returns the audit event flag
func (ae AuditEvent) Flag() Flag {
	return ae.flag
}

// WithTimestamp sets the message timestamp.
func (ae *AuditEvent) WithTimestamp(ts time.Time) *AuditEvent {
	ae.ts = ts
	return ae
}

// Timestamp returns the timed message timestamp.
func (ae AuditEvent) Timestamp() time.Time {
	return ae.ts
}

// WithHeading sets the heading.
func (ae *AuditEvent) WithHeading(heading string) *AuditEvent {
	ae.heading = heading
	return ae
}

// Heading returns the heading.
func (ae AuditEvent) Heading() string {
	return ae.heading
}

// WithPrincipal sets the principal.
func (ae *AuditEvent) WithPrincipal(principal string) *AuditEvent {
	ae.principal = principal
	return ae
}

// Principal returns the principal.
func (ae AuditEvent) Principal() string {
	return ae.principal
}

// WithVerb sets the verb.
func (ae *AuditEvent) WithVerb(verb string) *AuditEvent {
	ae.verb = verb
	return ae
}

// Verb returns the verb.
func (ae AuditEvent) Verb() string {
	return ae.verb
}

// WithNoun sets the noun.
func (ae *AuditEvent) WithNoun(noun string) *AuditEvent {
	ae.noun = noun
	return ae
}

// Noun returns the noun.
func (ae AuditEvent) Noun() string {
	return ae.noun
}

// WithSubject sets the subject.
func (ae *AuditEvent) WithSubject(subject string) *AuditEvent {
	ae.subject = subject
	return ae
}

// Subject returns the subject.
func (ae AuditEvent) Subject() string {
	return ae.subject
}

// WithProperty sets the property.
func (ae *AuditEvent) WithProperty(property string) *AuditEvent {
	ae.property = property
	return ae
}

// Property returns the property.
func (ae AuditEvent) Property() string {
	return ae.property
}

// WithExtra sets the extra info.
func (ae *AuditEvent) WithExtra(extra map[string]string) *AuditEvent {
	ae.extra = extra
	return ae
}

// Extra returns the extra information.
func (ae AuditEvent) Extra() map[string]string {
	return ae.extra
}

// WriteText implements TextWritable.
func (ae AuditEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	if len(ae.principal) > 0 {
		buf.WriteString(formatter.Colorize("Principal:", ColorGray))
		buf.WriteString(ae.principal)
		buf.WriteRune(RuneSpace)
	}
	if len(ae.verb) > 0 {
		buf.WriteString(formatter.Colorize("Verb:", ColorGray))
		buf.WriteString(ae.verb)
		buf.WriteRune(RuneSpace)
	}
	if len(ae.noun) > 0 {
		buf.WriteString(formatter.Colorize("Noun:", ColorGray))
		buf.WriteString(ae.noun)
		buf.WriteRune(RuneSpace)
	}
	if len(ae.subject) > 0 {
		buf.WriteString(formatter.Colorize("Subject:", ColorGray))
		buf.WriteString(ae.subject)
		buf.WriteRune(RuneSpace)
	}
	if len(ae.property) > 0 {
		buf.WriteString(formatter.Colorize("Property:", ColorGray))
		buf.WriteString(ae.property)
		buf.WriteRune(RuneSpace)
	}
	if len(ae.extra) > 0 {
		var values []string
		for key, value := range ae.extra {
			values = append(values, fmt.Sprintf("%s%s", formatter.Colorize(key+":", ColorGray), value))
		}
		buf.WriteString(strings.Join(values, " "))
	}
}

// WriteJSON implements JSONWritable.
func (ae AuditEvent) WriteJSON() JSONObj {
	return JSONObj{
		"principal": ae.principal,
		"verb":      ae.verb,
		"noun":      ae.noun,
		"subject":   ae.subject,
		"property":  ae.property,
		"extra":     ae.extra,
	}
}
