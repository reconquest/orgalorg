package main

type (
	callbackWriter func([]byte) (int, error)
)

func (writer callbackWriter) Write(data []byte) (int, error) {
	return writer(data)
}
