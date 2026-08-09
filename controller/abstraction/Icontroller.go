package abstraction

type Icontroller interface {
	Serve(socket string, ctrl *Icontroller) error
}
