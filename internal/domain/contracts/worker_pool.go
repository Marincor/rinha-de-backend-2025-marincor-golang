package contracts

type WorkerPoolManager interface {
	Submit(callback func())
	Wait()
}
