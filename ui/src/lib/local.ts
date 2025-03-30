export const getLocalItem = <T>(key: string): T | null => {
  const item = localStorage.getItem(key);
  return item ? JSON.parse(item) : null;
};

export const setLocalItem = <T>(key: string, value: T): void => {
  localStorage.setItem(key, JSON.stringify(value));
};
