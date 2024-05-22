package pkg

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const MQURL = "amqp://admin:admin@43.138.43.184:5673/admin"

type Rabbitmq struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机名称
	Exchange string
	// bindKey名称
	Key string
	//连接信息
	Mqurl string
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *Rabbitmq {
	return &Rabbitmq{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: MQURL}
}

// Destroy 断开channel和connection
func (r *Rabbitmq) Destroy() {
	r.conn.Close()
	r.channel.Close()
}

// 错误处理函数
func (r *Rabbitmq) failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		//panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func NewRabbitMQSimple(queueName string) *Rabbitmq {
	rabbitmq := NewRabbitMQ(queueName, "", "")
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "连接RabbitMQ失败")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "RabbitMQ创建管道失败")
	return rabbitmq
}

// PublishSimple Simple模式下队列生产者
func (r *Rabbitmq) PublishSimple(message string) {
	//1.申请队列，如果队列不存在则自动创建，存在则跳过
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	r.failOnErr(err, "声明发送的消息队列失败")

	//调用channel发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		//如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

}

// ConsumeSimple Simple模式下消费者
func (r *Rabbitmq) ConsumeSimple() {
	//1、申请队列如果队列不存在则自动创建，存在则跳过
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "声明接收的消息队列失败")

	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, //queue
		//用来区分多个消费者
		"", //consumer
		//是否自动应答
		true,
		//是否独有
		false,
		//设置为true，表示 不能将同一个Conenction中生产者发送的消息传递给这个Connection中 的消费者
		false,
		//列是否阻塞
		false,
		nil,
	)

	r.failOnErr(err, "创建消费者失败")
	//阻塞
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			//Func(string(d.Body))
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// PublishWork 工作模式下生产者:
func (r *Rabbitmq) PublishWork(message string) {
	//申请队列
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "声明工作模式队列失败")
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// ConsumeWork 工作模式下消费者：
func (r *Rabbitmq) ConsumeWork() {
	q, err := r.channel.QueueDeclare(r.QueueName, false, false, false, false, nil)
	r.failOnErr(err, "声明消费者工作队列失败")
	//*---------------------公平分发------------------*/
	err = r.channel.Qos(
		1,     //指定每个消费者能够同时处理的未确认消息的最大数量。例如，如果
		0,     //预取的消息大小，通常为 0，表示未指定
		false, //表示这些参数是针对信道的全局设置还是针对每个消费者的独立设置。如果 global 为 true，则表示这些参数将应用于所有消费者；如果 global 为 false，则表示这些参数将仅应用于当前的信道。
	)
	msgs, err := r.channel.Consume(q.Name, "", true, false, false, false, nil)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	//阻塞
	select {}
}

// NewRabbitMQPubSub 订阅模式下创建RabbitMQ实例
func NewRabbitMQPubSub(exchangeName string) *Rabbitmq {
	rabbitmq := NewRabbitMQ("", exchangeName, "")
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "订阅模式建立连接失败")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "订阅模式创建channel失败")
	return rabbitmq
}

// PublishPub 订阅模式下生产者
func (r *Rabbitmq) PublishPub(message string) {
	//1、创建交换机
	//err := r.channel.QueueDeclare()
	err := r.channel.ExchangeDeclare(
		r.Exchange, //交换机名字
		"topic",    //交换机类型
		true,       //是否持久化
		false,      //是否自动删除，表示交换机在没有与之绑定的队列时是否自动删除
		false,      // 是否是内部交换机，表示该交换机是否被客户端使用
		false,      // 是否等待服务器响应，这里设置为 false。
		nil,
	)
	r.failOnErr(err, "创建交换机失败")

	err = r.channel.Publish(
		r.Exchange, // 表示将消息发送到默认的交换机
		r.Key,      // 指定了消息的路由键，即将消息发送到的队列
		false,      //是 它表示如果无法将消息路由到队列中， 消息会被丢弃而不是返回给发送者。
		false,      //是 它表示如果无法立即将消息发送给接收者， 消息会被丢弃而不是返回给发送者
		amqp.Publishing{ContentType: "text/plain", Body: []byte(message)})
}

// ConsumePub 订阅模式消费者
func (r *Rabbitmq) ConsumePub() {
	//创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange, //交换机名字
		"topic",    //交换机类型
		true,       //是否持久化
		false,      //是否自动删除，表示交换机在没有与之绑定的队列时是否自动删除
		false,      // 是否是内部交换机，表示该交换机是否被客户端使用
		false,      // 是否等待服务器响应，这里设置为 false。
		nil,
	)
	r.failOnErr(err, "消费者交换机声明失败")

	q, err := r.channel.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "消费者随机队列声明失败")

	//绑定队列到exchange中
	err = r.channel.QueueBind(
		q.Name,
		"", //在pub/sub模式下，这里的key要为空
		r.Exchange,
		false,
		nil,
	)
	r.failOnErr(err, "绑定失败")

	//消费信息
	message, err := r.channel.Consume(
		q.Name, //表示要消费的队列名称
		"",     //表示消费者的名称，这里为空字符串。
		true,   //表示开启自动确认模式，即当消费者接收到消息后自动向 RabbitMQ 发送消息已被消费的回执(消息确认)
		false,  //表示不加锁。
		false,  //表示不禁止消费者使用同一连接发送消息。
		false,  //表示不等待接收返回的参数。
		nil,    //表示不传递其他参数。
	)
	r.failOnErr(err, "创建消费者队列失败")

	go func() {
		for d := range message {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	fmt.Println("退出请按 CTRL+C")
	select {}
}

// NewRabbitMQRouting 路由模式创建RabbitMQ实例
func NewRabbitMQRouting(exchangeName string, routingKey string) *Rabbitmq {
	rabbitmq := NewRabbitMQ("", exchangeName, routingKey)
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建路由模式实例失败")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "创建管道失败")
	return rabbitmq
}

// PublishRouting 路由模式生产者
func (r *Rabbitmq) PublishRouting(message string) {
	//创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange, //交换机名字
		"direct",   //交换机类型
		true,       //是否持久化
		false,      //是否自动删除，表示交换机在没有与之绑定的队列时是否自动删除
		false,      // 是否是内部交换机，表示该交换机是否被客户端使用
		false,      // 是否等待服务器响应，这里设置为 false。
		nil,
	)
	r.failOnErr(err, "创建交换机失败")

	err = r.channel.Publish(
		r.Exchange,
		r.Key, //路由
		false,
		false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(message)},
	)
}

// ConsumeRouting 路由模式消费者
func (r *Rabbitmq) ConsumeRouting() {
	err := r.channel.ExchangeDeclare(
		r.Exchange, //交换机名字
		"direct",   //交换机类型
		true,       //是否持久化
		false,      //是否自动删除，表示交换机在没有与之绑定的队列时是否自动删除
		false,      // 是否是内部交换机，表示该交换机是否被客户端使用
		false,      // 是否等待服务器响应，这里设置为 false。
		nil,
	)
	r.failOnErr(err, "路由模式交换机声明失败")

	q, err := r.channel.QueueDeclare(
		"",    // name
		false, // 持久的
		false, //当队列不再使用时是否自动删除
		true,  //是否设置排他性队列（只能由声明它的连接使用）
		false, //是否设置为无等待（no-wait，等待服务器响应）。
		nil,
	)
	r.failOnErr(err, "创建队列失败")

	err = r.channel.QueueBind(
		q.Name,
		r.Key,
		r.Exchange,
		false,
		nil,
	)

	msgs, err := r.channel.Consume(
		q.Name, //表示要消费的队列名称
		"",     //表示消费者的名称，这里为空字符串。
		true,   //表示开启自动确认模式，即当消费者接收到消息后自动向 RabbitMQ 发送消息已被消费的回执(消息确认)
		false,  //表示不加锁。
		false,  //表示不禁止消费者使用同一连接发送消息。
		false,  //表示不等待接收返回的参数。
		nil,    //表示不传递其他参数。
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	select {}
}

// NewRabbitMQTopic Topic模式
func NewRabbitMQTopic(exchangeName string, routingKey string) *Rabbitmq {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ("", exchangeName, routingKey)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// 话题模式发送消息
func (r *Rabbitmq) PublishTopic(message string) {
	//1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		//要改成topic
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	//2.发送消息
	err = r.channel.Publish(
		r.Exchange,
		//要设置
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// RecieveTopic 匹配 kuteng.* 表示匹配 kuteng.hello, kuteng.hello.one需要用kuteng.#才能匹配到
func (r *Rabbitmq) ConsumeTopic() {
	//1.试探性创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		//交换机类型
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")
	//2.试探性创建队列，这里注意队列名称不要写
	q, err := r.channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	//绑定队列到 exchange 中
	err = r.channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		r.Key,
		r.Exchange,
		false,
		nil)

	//消费消息
	messges, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range messges {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	select {}
}
