# go-musthave-shortener-tpl

Сервис сокращения URL -

### Начало работы:

Смотреть Makefile

## DEBUG

### Профилирование памяти (pprof)

**Тестировалось через Postman 13 методов одновременно каждые 20ms 100 iteration**

```bash
$ go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

File: main
Type: inuse_space
Time: 2026-07-13 00:23:11 MSK
Showing nodes accounting for -1026.01kB, 40.02% of 2564.05kB total
      flat  flat%   sum%        cum   cum%
   -1026kB 40.01% 40.01%    -1026kB 40.01%  runtime.allocm
 -512.05kB 19.97% 59.99%  -512.05kB 19.97%  runtime.(*scavengerState).init
  512.04kB 19.97% 40.02%   512.04kB 19.97%  context.withCancel (inline)
         0     0% 40.02%   512.04kB 19.97%  context.WithCancel
         0     0% 40.02%   512.04kB 19.97%  github.com/jackc/pgx/v5/pgxpool.(*Pool).createIdleResources
         0     0% 40.02%   512.04kB 19.97%  github.com/jackc/pgx/v5/pgxpool.NewWithConfig.func5
         0     0% 40.02%  -512.05kB 19.97%  runtime.bgscavenge
         0     0% 40.02%    -1026kB 40.01%  runtime.mstart
         0     0% 40.02%    -1026kB 40.01%  runtime.mstart0
         0     0% 40.02%    -1026kB 40.01%  runtime.mstart1
         0     0% 40.02%    -1026kB 40.01%  runtime.newm
         0     0% 40.02%    -1026kB 40.01%  runtime.resetspinning
         0     0% 40.02%    -1026kB 40.01%  runtime.schedule
         0     0% 40.02%    -1026kB 40.01%  runtime.startm
         0     0% 40.02%    -1026kB 40.01%  runtime.wakep
```

### Benchmark

**BatchSave**

```bash
BenchmarkBatchSave/Size_1-10             2461420             464.6 ns/op             256 B/op          9 allocs/op
BenchmarkBatchSave/Size_100-10             26619             45171 ns/op           31712 B/op        808 allocs/op
BenchmarkBatchSave/Size_1000-10             2608            448701 ns/op          294242 B/op       8011 allocs/op
BenchmarkBatchSave/Size_10000-10             258           4635171 ns/op         3711093 B/op      80019 allocs/op

# Optimized (*pgx.Batch)
BenchmarkBatchSave/Size_1-10             3331099               356.5 ns/op           168 B/op          5 allocs/op
BenchmarkBatchSave/Size_10-10            3342674               355.6 ns/op           168 B/op          5 allocs/op
BenchmarkBatchSave/Size_100-10           3387565               355.0 ns/op           168 B/op          5 allocs/op
BenchmarkBatchSave/Size_1000-10          3338485               356.1 ns/op           168 B/op          5 allocs/op
BenchmarkBatchSave/Size_10000-10         3211534               364.5 ns/op           168 B/op          5 allocs/op
```

**Audit**

```bash
BenchmarkBufferSizes/buf_0-10            5763944               190.3 ns/op            32 B/op          1 allocs/op
BenchmarkBufferSizes/buf_1-10            8271272               144.9 ns/op            32 B/op          1 allocs/op
BenchmarkBufferSizes/buf_10-10          12249568                98.14 ns/op           32 B/op          1 allocs/op
BenchmarkBufferSizes/buf_20-10          12122236                98.23 ns/op           32 B/op          1 allocs/op
BenchmarkBufferSizes/buf_100-10         12386829               102.0 ns/op            32 B/op          1 allocs/op

# С ростом задержки в Notify (например, из-за медленного внешнего сервиса)
# буфер канала быстро заполняется, и AddLog начинает блокироваться
BenchmarkSlowSubscriber/delay_0s-10     12058078             98.95 ns/op              32 B/op          1 allocs/op
BenchmarkSlowSubscriber/delay_1µs-10      292279              4088 ns/op              32 B/op          1 allocs/op
BenchmarkSlowSubscriber/delay_10µs-10      79135             14571 ns/op              32 B/op          1 allocs/op

BenchmarkParallel-10                     5070506               238.1 ns/op            32 B/op          1 allocs/op

# buf size 20
BenchmarkSlowHTTPObserver-10                 262           5305591 ns/op            6520 B/op         68 allocsop
# buf size 50
BenchmarkSlowHTTPObserver-10                 423           5067079 ns/op            5935 B/op         65 allocs/op
```

**Gzip**

```bash
# No pool
BenchmarkGzipCompress/gzip_alloc-10        12577             95876 ns/op         1722411 B/op         78 allocs/op
# compressWriter pool+
BenchmarkGzipCompress/gzip_alloc-10        25459             46193 ns/op          908102 B/op         60 allocs/op
# compressReader pool+
BenchmarkGzipCompress/gzip_alloc-10        23913             49226 ns/op          871062 B/op         55 allocs/op
```

### Оптимизация

Были проведены следующие оптимизации:

- Оптимизирован gzip middleware: добавлен sync.Pool для переиспользования gzip.Writer
- Улучшена работа с пулом соединений PostgreSQL
