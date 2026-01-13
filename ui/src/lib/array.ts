export type Entry<K, V> = {
  key: K;
  value: V;
};

export const toEntries = <K extends string, T>(
  record: Partial<Record<K, T>>
): Entry<K, T>[] =>
  Object.entries(record).map(([key, value]) => ({
    key: key as K,
    value: value as T,
  }));

export const recordMap = <K extends string | number, T, V>(
  record: Record<K, T>,
  fn: (value: T, key: K) => V
): Record<K, V> => {
  return Object.fromEntries(
    Object.entries(record).map(([key, value]) => [
      key,
      fn(value as T, key as K),
    ])
  ) as Record<K, V>;
};

export const groupItems = <T, K extends string | number>(
  items: T[],
  getKey: (item: T) => K
): Record<K, T[]> => {
  const result = {} as Record<K, T[]>;

  items.forEach((item) => {
    const key = getKey(item);
    if (!result[key]) {
      result[key] = [];
    }
    result[key]?.push(item);
  });

  return result;
};

export type Order = "asc" | "desc";

export const order = <T>(
  array: readonly T[],
  getValue: (item: T) => number,
  order: Order
) => {
  return [...array].sort((a, b) => {
    if (order === "asc") {
      return getValue(a) - getValue(b);
    } else {
      return getValue(b) - getValue(a);
    }
  });
};
