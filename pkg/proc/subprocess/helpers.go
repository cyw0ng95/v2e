package subprocess

// NewErrorResponse creates an error response message from a request message.
// It sets Type to MessageTypeError, copies the ID and CorrelationID from the original message,
// sets the Error field to errMsg, and sets Target to the original message's Source.
func NewErrorResponse(msg *Message, errMsg string) *Message {
	return &Message{
		Type:          MessageTypeError,
		ID:            msg.ID,
		Error:         errMsg,
		CorrelationID: msg.CorrelationID,
		Target:        msg.Source,
	}
}

// NewErrorResponseWithPrefix creates an error response message with a service prefix.
// It is identical to NewErrorResponse but prepends "[prefix] " to the error message.
// This is useful for identifying which service generated the error.
func NewErrorResponseWithPrefix(msg *Message, prefix, errMsg string) *Message {
	prefixedErrMsg := "[" + prefix + "] " + errMsg
	return &Message{
		Type:          MessageTypeError,
		ID:            msg.ID,
		Error:         prefixedErrMsg,
		CorrelationID: msg.CorrelationID,
		Target:        msg.Source,
	}
}

// NewSuccessResponse creates a success response message from a request message.
// It sets Type to MessageTypeResponse, copies the ID and CorrelationID from the original message,
// marshals the result as the Payload, sets Target to the original message's Source,
// and sets Source to the original message's Target.
// Returns an error if marshaling the result fails.
func NewSuccessResponse(msg *Message, result interface{}) (*Message, error) {
	response := &Message{
		Type:          MessageTypeResponse,
		ID:            msg.ID,
		CorrelationID: msg.CorrelationID,
		Target:        msg.Source,
		Source:        msg.Target,
	}

	if result != nil {
		payload, err := MarshalFast(result)
		if err != nil {
			return nil, err
		}
		response.Payload = payload
	}

	return response, nil
}

// IsErrorResponse checks if a message is an error response.
// Returns (true, msg.Error) if the message type is MessageTypeError.
// Returns (false, "") for all other message types, including nil messages.
func IsErrorResponse(msg *Message) (bool, string) {
	if msg == nil {
		return false, ""
	}
	if msg.Type == MessageTypeError {
		return true, msg.Error
	}
	return false, ""
}
