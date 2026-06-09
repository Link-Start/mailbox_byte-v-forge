package natseventbus

import commonv1 "mailboxapi/internal/contracts/commonv1"

type deadLetterSource struct {
	id            string
	name          string
	version       string
	source        string
	correlationID string
	traceID       string
}

func deadLetterOriginal(envelope *commonv1.EventEnvelope) deadLetterSource {
	metadata := envelope.GetMetadata()
	if metadata == nil {
		return deadLetterSource{}
	}
	return deadLetterSource{
		id:            metadata.GetId(),
		name:          metadata.GetType(),
		version:       metadata.GetVersion(),
		source:        metadata.GetSource(),
		correlationID: metadata.GetCorrelationId(),
		traceID:       metadata.GetTraceId(),
	}
}
