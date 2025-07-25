import { useState } from "react";

export function useLocalStorage<T>(
  keyName: string,
  defaultValue: T
): [T, (value: T) => void] {
  const [storedValue, setStoredValue] = useState(() => {
    try {
      const value = window.localStorage.getItem(keyName);
      if (value) {
        return JSON.parse(value) as T;
      } else {
        window.localStorage.setItem(keyName, JSON.stringify(defaultValue));
        return defaultValue;
      }
    } catch (err) {
      console.log(err);
      return defaultValue;
    }
  });
  const setValue = (newValue: T) => {
    try {
      window.localStorage.setItem(keyName, JSON.stringify(newValue));
    } catch (err) {
      console.log(err);
    }
    setStoredValue(newValue);
  };
  return [storedValue, setValue];
}

export function useNullableLocalStorage<T>(
  keyName: string,
  defaultValue: T | null
): [T | null, (value: T | null) => void] {
  const [storedValue, setStoredValue] = useState(() => {
    try {
      const value = window.localStorage.getItem(keyName);
      if (value) {
        return JSON.parse(value) as T;
      } else {
        if (!defaultValue) {
          return defaultValue;
        } else {
          window.localStorage.setItem(keyName, JSON.stringify(defaultValue));
          return defaultValue;
        }
      }
    } catch (err) {
      console.error(err);
      return defaultValue;
    }
  });
  const setValue = (newValue: T | null) => {
    try {
      if (!newValue) {
        window.localStorage.removeItem(keyName);
      } else {
        window.localStorage.setItem(keyName, JSON.stringify(newValue));
      }
    } catch (err) {
      console.log(err);
    }
    setStoredValue(newValue);
  };
  return [storedValue, setValue];
}
