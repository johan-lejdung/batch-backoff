# batch-backoff
Provides exponential backoff for batch processes

Useful when you want exponential backoff when running async batch processes. eg. when you want the following to only result in a single backoff operation.

```
for i := 0; i < 20; i++ {
    doProcessCall()

    Backoff()
}
```

Create a new instance of the struct with:
```
backoff := NewExponentialBackoff(BackoffIntervals{
    StartInterval: 10 * time.Minute,
    Multiplier:    2,
    MaxInterval:   2 * time.Minute,
})
```

You use the library by first calling `CanProceed()`.

You will get back `(bool, BackoffBatch)`, where the bool indicates if you can proceed or not.


The batch is used when calling `Backoff(batch BackoffBatch)`. If the call made after `CanProceed()` fails you just need to provice the `BackoffBatch` to the call of `Backoff` and the increments will handle it self.
