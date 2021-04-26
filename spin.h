#ifndef SYNC_H
#define SYNC_H

#include <stdatomic.h>
#include <stdbool.h>

/* type alias for spinlock */
typedef atomic_bool spinlock;

/* Lock locks l.
 * If the lock is already in use, the calling thread
 * blocks until the locker is available.
 */
void lock(spinlock* l);

/* try_lock should first check if the lock is free
 * before attempting to acquire it.
 * This would prevent excessive coherency
 * traffic in case someone loops over try_lock().
 */
bool try_lock(spinlock*);

/* Unlock unlocks l. */
void unlock(spinlock*);

#endif /* SYNC_H */
