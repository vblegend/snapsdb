module snapsdb

go 1.18

require (
	github.com/BurntSushi/toml v1.2.0 //toml 配置文档支持
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 //  swagger文档模板
	github.com/bitly/go-simplejson v0.5.0 // json 序列化加强版
	github.com/bytedance/go-tagexpr/v2 v2.9.5 // API 参数校验
	github.com/casbin/casbin/v2 v2.51.2 // 权限框架
	github.com/fsnotify/fsnotify v1.5.4 // 文件变化监控库
	github.com/ghodss/yaml v1.0.0 // yaml 对象序列化
	github.com/gin-gonic/gin v1.8.1 // gin web框架
	github.com/go-redis/redis/v7 v7.4.1 // redis sdk
	github.com/google/uuid v1.3.0 // uuid
	github.com/imdario/mergo v0.3.13 // json 合并库
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mssola/user_agent v0.5.3 // 浏览器 user_agent 分析工具
	github.com/opentracing/opentracing-go v1.2.0 // API链路追踪库
	github.com/pkg/errors v0.9.1 // golang errors 辅助库
	github.com/prometheus/client_golang v1.13.0 // 普罗米修斯 客户端接口
	github.com/robfig/cron/v3 v3.0.1 // cron 任务调度引擎
	github.com/shirou/gopsutil/v3 v3.22.7 // 计算机硬件资源采集库
	github.com/spf13/cast v1.5.0 // 数据转换 格式化库
	github.com/spf13/cobra v1.5.0 // CIL命令行框架
	github.com/swaggo/gin-swagger v1.5.2 // swagger API文档
	github.com/swaggo/swag v1.8.4 // swagger API文档
	github.com/unrolled/secure v1.12.0 // gin ssl https
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // 哈希加解密库
	gorm.io/driver/mysql v1.3.6 // mysql 连接库
	gorm.io/driver/sqlite v1.3.6 // sqlite 连接库
	gorm.io/gorm v1.23.8 // go orm框架
	gorm.io/plugin/dbresolver v1.2.2 // orm 多数据库支持

)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // jwt 登录
	github.com/go-playground/locales v0.14.0 // 多语言支持
	github.com/go-playground/universal-translator v0.18.0 // 多语言支持
	github.com/go-playground/validator/v10 v10.11.0
	github.com/gorilla/websocket v1.5.0 // websocket 支持
	github.com/shamsher31/goimgext v1.0.0 // 图片扩展名
)

require (
	github.com/docker/docker v20.10.17+incompatible // docker sdk
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/nakabonne/tstorage v0.3.5 // 时序数据库
	// golang.org/x/net v0.0.0-20220615171555-694bf12d69de // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	gotest.tools/v3 v3.3.0 // indirect
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/andeya/goutil v0.0.0-20220704075712-42f2ec55fe8d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-openapi/swag v0.22.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/goccy/go-json v0.9.10 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/henrylee2cn/ameda v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20220517141722-cf486979b281 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nyaruka/phonenumbers v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.3 // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/swaggo/files v0.0.0-20220728132757-551d4a08d97a
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.5.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/gin-contrib/pprof v1.4.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.18.1 // indirect
)
