#include "spin.h"

void lock(spinlock* lock)
{
    for (;;) {
	/* Optimistically assume the lock is free on the first try */
	if (!atomic_exchange_explicit(lock, true, memory_order_acquire)) {
	    return;
	}

	/* Wait for lock to be released without generating cache misses */
	while (atomic_load_explicit(lock, memory_order_relaxed)) {
	    /* Issue X86 PAUSE or ARM YIELD instruction to reduce contention between
	     * Simultaneous Multithreading. */
	    __builtin_ia32_pause();
	}
    }
}

bool try_lock(spinlock* lock)
{
    /* First do a relaxed load to check if lock is free in order to prevent
     * unnecessary cache misses if someone does while(!try_lock()) */
    return !atomic_load_explicit(lock, memory_order_relaxed)
	&& !atomic_exchange_explicit(lock, true, memory_order_acquire);
}

void unlock(spinlock* lock)
{
    atomic_store_explicit(lock, false, memory_order_release);
}
