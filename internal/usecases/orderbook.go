package usecases

import (
	"context"
	"fmt"
	"sync"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"

	"github.com/bataloff/tiknkoff/domain"
)

type OrderBookRepository interface {
	Store(ctx context.Context, orders []*domain.OrderBook) error
}

type OrderBookUseCase struct {
	orderBookRepository OrderBookRepository

	investClient *investgo.Client
	instruments  *investgo.InstrumentsServiceClient

	wg *sync.WaitGroup
}

func NewOrderBookUseCase(
	ob OrderBookRepository,
	client *investgo.Client,
) *OrderBookUseCase {
	return &OrderBookUseCase{
		orderBookRepository: ob,
		investClient:        client,
		instruments:         client.NewInstrumentsServiceClient(),
		wg:                  &sync.WaitGroup{},
	}
}

func (o *OrderBookUseCase) Ping(ctx context.Context) error {
	response, err := o.investClient.NewUsersServiceClient().GetAccounts()
	if err != nil {
		return err
	}
	accounts := response.GetAccounts()

	o.investClient.Logger.Infof("Ping Tinkoff")
	for _, account := range accounts {
		o.investClient.Logger.Infof(" ==> account id = %v\n", account.GetId())
	}
	return nil
}

// Sync данный метод можно разделить на множество отдельных методов, чтобы не было такой большой конгнитивнйо нагрузки
func (o *OrderBookUseCase) Sync(ctx context.Context) error {
	const (
		BatchSize = 300
		Depth     = 20
	)

	// получаем список акций доступных для торговли через investAPI
	instrumentsResp, err := o.instruments.Shares(pb.InstrumentStatus_INSTRUMENT_STATUS_BASE)
	if err != nil {
		return err
	}
	// слайс идентификаторов торговых инструментов
	instrumentIds := make([]string, 0, 900)

	// берем первые 900 элементов
	instruments := instrumentsResp.GetInstruments()
	for i, instrument := range instruments {
		if i > 899 {
			break
		}
		instrumentIds = append(instrumentIds, instrument.GetUid())
	}

	fmt.Printf("got %v instruments\n", len(instrumentIds))

	// создаем клиента сервиса стримов маркетдаты, и с его помощью создаем стримы
	MarketDataStreamService := o.investClient.NewMarketDataStreamClient()
	// создаем стримы
	stream1, err := MarketDataStreamService.MarketDataStream()
	if err != nil {
		return err
	}

	stream2, err := MarketDataStreamService.MarketDataStream()
	if err != nil {
		return err
	}

	stream3, err := MarketDataStreamService.MarketDataStream()
	if err != nil {
		return err
	}
	// в рамках каждого стрима подписываемся на стаканы для 300 инструментов
	orderBooks1, err := stream1.SubscribeOrderBook(instrumentIds[:299], Depth)
	if err != nil {
		return err
	}

	orderBooks2, err := stream2.SubscribeOrderBook(instrumentIds[300:599], Depth)
	if err != nil {
		return err
	}

	orderBooks3, err := stream3.SubscribeOrderBook(instrumentIds[600:899], Depth)
	if err != nil {
		return err
	}
	// процесс сохранения стакана в хранилище:
	// чтение из стрима -> преобразование -> сохранение в слайс -> запись в хранилище
	// разбиваем процесс на три горутины:
	// 1. Читаем из стрима и преобразуем -> orderBookStorage
	// 2. Собираем batch и отправляем во внешнее хранилище -> externalStorage
	// 3. Записываем batch в хранилище

	// канал для связи горутны 1 и 2
	orderBookStorage := make(chan *domain.OrderBook)
	defer close(orderBookStorage)

	// запускаем чтение стримов
	o.wg.Add(1)
	go func(ctx context.Context) {
		defer o.wg.Done()
		err := stream1.Listen()
		if err != nil {
			o.investClient.Logger.Errorf(err.Error())
		}
	}(ctx)

	o.wg.Add(1)
	go func(ctx context.Context) {
		defer o.wg.Done()
		err := stream2.Listen()
		if err != nil {
			o.investClient.Logger.Errorf(err.Error())
		}
	}(ctx)

	o.wg.Add(1)
	go func(ctx context.Context) {
		defer o.wg.Done()
		err := stream3.Listen()
		if err != nil {
			o.investClient.Logger.Errorf(err.Error())
		}
	}(ctx)

	// читаем стаканы из каналов и преобразуем стакан в нужную структуру
	o.wg.Add(1)
	go func(ctx context.Context) {
		defer func() {
			// если мы слушаем в одной рутине несколько стримов, то
			// при завершении (из-за закрытия одного из каналов) нужно остановить все стримы
			stream1.Stop()
			stream2.Stop()
			stream3.Stop()
			o.wg.Done()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case input, ok := <-orderBooks1:
				if !ok {
					return
				}
				orderBookStorage <- transformOrderBook(input)
			case input, ok := <-orderBooks2:
				if !ok {
					return
				}
				orderBookStorage <- transformOrderBook(input)
			case input, ok := <-orderBooks3:
				if !ok {
					return
				}
				orderBookStorage <- transformOrderBook(input)
			}
		}
	}(ctx)

	// канал для связи горутины 2 и 3
	externalStorage := make(chan []*domain.OrderBook)
	defer close(externalStorage)
	// сохраняем в хранилище
	o.wg.Add(1)
	go func(ctx context.Context) {
		defer o.wg.Done()

		batch := make([]*domain.OrderBook, BatchSize)
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case ob, ok := <-orderBookStorage:
				if !ok {
					return
				}
				if count < BatchSize-1 {
					batch[count] = ob
					count++
				} else {
					o.investClient.Logger.Infof("batch overflow")
					batch[count] = ob
					count = 0
					externalStorage <- batch
				}
			}
		}
	}(ctx)

	// записываем стаканы в бд или json
	o.wg.Add(1)
	go func(ctx context.Context) {
		defer o.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-externalStorage:
				if !ok {
					return
				}

				if err = o.orderBookRepository.Store(ctx, data); err != nil {
					o.investClient.Logger.Errorf(err.Error())
				}
			}
		}
	}(ctx)

	o.wg.Wait()

	return nil
}

// transformOrderBook - Преобразование стакана в нужный формат
func transformOrderBook(input *pb.OrderBook) *domain.OrderBook {
	depth := input.GetDepth()
	bids := make([]domain.Order, 0, depth)
	asks := make([]domain.Order, 0, depth)
	for _, o := range input.GetBids() {
		bids = append(bids, domain.Order{
			Price:    o.GetPrice().ToFloat(),
			Quantity: o.GetQuantity(),
		})
	}
	for _, o := range input.GetAsks() {
		asks = append(asks, domain.Order{
			Price:    o.GetPrice().ToFloat(),
			Quantity: o.GetQuantity(),
		})
	}
	return &domain.OrderBook{
		Figi:          input.GetFigi(),
		InstrumentUid: input.GetInstrumentUid(),
		Depth:         depth,
		IsConsistent:  input.GetIsConsistent(),
		TimeUnix:      input.GetTime().AsTime().Unix(),
		LimitUp:       input.GetLimitUp().ToFloat(),
		LimitDown:     input.GetLimitDown().ToFloat(),
		Bids:          bids,
		Asks:          asks,
	}
}
