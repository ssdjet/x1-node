package processor_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygonHermez/zkevm-node/etherman"
	"github.com/0xPolygonHermez/zkevm-node/synchronizer/actions"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

type ProcessorStub struct {
	name             string
	supportedEvents  []etherman.EventOrder
	supportedForkIds []actions.ForkIdType
	responseProcess  error
}

func (p *ProcessorStub) Name() string {
	return p.name
}

func (p *ProcessorStub) SupportedEvents() []etherman.EventOrder {
	return p.supportedEvents
}

func (p *ProcessorStub) SupportedForkIds() []actions.ForkIdType {
	return p.supportedForkIds
}

func (p *ProcessorStub) Process(ctx context.Context, order etherman.Order, l1Block *etherman.Block, dbTx pgx.Tx) error {
	return p.responseProcess
}

func TestL1EventProcessors_Get(t *testing.T) {
	// Create a new instance of L1EventProcessors

	// Create some test data
	forkId1 := actions.ForkIdType(1)
	forkId2 := actions.ForkIdType(2)
	event1 := etherman.EventOrder("event1")
	event2 := etherman.EventOrder("event2")
	processorConcrete := ProcessorStub{
		name:             "processor_event1_forkid1",
		supportedEvents:  []etherman.EventOrder{event1},
		supportedForkIds: []actions.ForkIdType{forkId1},
		responseProcess:  nil,
	}
	processorConcreteForkId2 := ProcessorStub{
		name:             "processor_event2_forkid2",
		supportedEvents:  []etherman.EventOrder{event2},
		supportedForkIds: []actions.ForkIdType{forkId2},
		responseProcess:  nil,
	}
	processorWildcard := ProcessorStub{
		name:             "processor_event1_forkidWildcard",
		supportedEvents:  []etherman.EventOrder{event1},
		supportedForkIds: []actions.ForkIdType{actions.WildcardForkId},
		responseProcess:  nil,
	}
	builder := NewL1EventProcessorsBuilder()
	builder.Register(&processorConcrete)
	builder.Register(&processorWildcard)
	builder.Register(&processorConcreteForkId2)
	sut := builder.Build()

	result := sut.Get(forkId1, event1)
	require.Equal(t, &processorConcrete, result, "must return concrete processor")
	result = sut.Get(forkId2, event1)
	require.Equal(t, &processorWildcard, result, "must return wildcard processor")
	result = sut.Get(forkId1, event2)
	require.Equal(t, nil, result, "no processor")
}

func TestL1EventProcessors_Process(t *testing.T) {
	forkId1 := actions.ForkIdType(1)

	event1 := etherman.EventOrder("event1")
	event2 := etherman.EventOrder("event2")

	processorConcrete := ProcessorStub{
		name:             "processor_event1_forkid1",
		supportedEvents:  []etherman.EventOrder{event1},
		supportedForkIds: []actions.ForkIdType{forkId1},
		responseProcess:  nil,
	}
	processorConcreteEvent2 := ProcessorStub{
		name:             "processor_event1_forkid1",
		supportedEvents:  []etherman.EventOrder{event2},
		supportedForkIds: []actions.ForkIdType{forkId1},
		responseProcess:  errors.New("error2"),
	}
	builder := NewL1EventProcessorsBuilder()
	builder.Register(&processorConcrete)
	builder.Register(&processorConcreteEvent2)
	sut := builder.Build()

	result := sut.Process(context.Background(), forkId1, etherman.Order{Name: event1, Pos: 0}, nil, nil)
	require.Equal(t, processorConcrete.responseProcess, result, "must return concrete processor response")

	result = sut.Process(context.Background(), forkId1, etherman.Order{Name: event2, Pos: 0}, nil, nil)
	require.Equal(t, processorConcreteEvent2.responseProcess, result, "must return concrete processor response")

	result = sut.Process(context.Background(), actions.ForkIdType(2), etherman.Order{Name: event1, Pos: 0}, nil, nil)
	require.ErrorIs(t, result, ErrCantProcessThisEvent, "must return not found error")
}
