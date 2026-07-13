## Профилирование памяти (pprof)

### Оптимизация

Были проведены следующие оптимизации:

- Оптимизирован gzip middleware: добавлен sync.Pool для переиспользования gzip.Writer
- Улучшена работа с пулом соединений PostgreSQL

**Тестировалось через Postman 13 методов каждые 20ms 100 iteration**

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
