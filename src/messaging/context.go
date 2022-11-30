package messaging

type MessageContext interface {
	Headers() map[string]string
	Message() interface{}
}

func NewMessageContext(headers map[string]string, message interface{}) MessageContext {
	return &messageContext{
		message: message,
		headers: headers,
	}
}

type messageContext struct {
	message interface{}
	headers map[string]string
}

func (c *messageContext) Headers() map[string]string {
	return c.headers
}

func (c *messageContext) Message() interface{} {
	return c.message
}
