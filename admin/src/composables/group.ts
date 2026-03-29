import { computed } from 'vue'
import type { Ref } from 'vue'

export interface Group<T> {
  key: string
  items: T[]
}

/**
 * Groups a reactive array by a given key.
 * Returns `groupBy` as a computed array of `{ key, items }` objects so that
 * callers can safely use `.map()`, `.filter()`, etc. without hitting
 * "TypeError: groupBy.value.map is not a function" (which occurs when the
 * grouped result is returned as a plain object instead of an array).
 */
export function useGroupBy<T extends Record<string, unknown>>(
  items: Ref<T[]>,
  key: keyof T
) {
  const groupBy = computed<Group<T>[]>(() => {
    const map = new Map<string, T[]>()

    for (const item of items.value) {
      const groupKey = String(item[key] ?? '')
      if (!map.has(groupKey)) {
        map.set(groupKey, [])
      }
      map.get(groupKey)!.push(item)
    }

    // Return an array so callers can safely call .map(), .filter(), etc.
    return Array.from(map.entries()).map(([k, groupItems]) => ({
      key: k,
      items: groupItems
    }))
  })

  return { groupBy }
}
