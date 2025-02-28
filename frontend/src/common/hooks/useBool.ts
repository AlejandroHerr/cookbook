import { useCallback, useState } from 'react';

import { VoidCB } from '@/types';

interface UseBool {
  value: boolean;
  setTrue: VoidCB;
  setFalse: VoidCB;
  toggle: VoidCB;
}

export const useBool = (initialValue?: boolean): UseBool => {
  const [value, setValue] = useState(!!initialValue);

  const setTrue = useCallback(() => {
    setValue(true);
  }, []);
  const setFalse = useCallback(() => {
    setValue(false);
  }, []);
  const toggle = useCallback(() => {
    setValue((prev) => !prev);
  }, []);

  return { value, setTrue, setFalse, toggle };
};
