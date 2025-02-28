import { Command as CommandPrimitive } from 'cmdk';
import React, { useEffect, useState } from 'react';

import { Command, CommandGroup, CommandItem, CommandList } from './command';

import { useBool } from '@/common/hooks/useBool';
import { cn } from '@/lib/utils';

type ASOption = string;
interface AsyncSelectProps {
  onSearch: (search: string) => Promise<ASOption[]>;
  onChange: (value: ASOption) => void;
  commandProps?: React.ComponentPropsWithoutRef<typeof Command>;
  /** Props of `CommandInput` */
  inputProps?: Omit<
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input>,
    'value' | 'placeholder' | 'disabled'
  >;
  loadingIndicator?: React.ReactNode;
  triggerSearchOnFocus?: boolean;
  placeholder?: string;
  value?: ASOption;
  disabled?: boolean;
  className?: string;
}

export interface AsyncSelectRef {
  selectedValue: ASOption;
  input: HTMLInputElement;
  focus: () => void;
  reset: () => void;
}

export const AsyncSelect = React.forwardRef<AsyncSelectRef, AsyncSelectProps>(
  ({
    commandProps,
    inputProps,
    onChange,
    value,
    disabled,
    className,
    placeholder,
    onSearch,
  }) => {
    const dropdownRef = React.useRef<HTMLDivElement>(null); // Added this
    const inputRef = React.useRef<HTMLInputElement>(null);
    const {
      value: isLoading,
      setTrue: setIsLoading,
      setFalse: unsetIsLoading,
    } = useBool();
    const [onScrollbar, setOnScrollbar] = useState(false);
    const [open, setOpen] = useState(false);
    const [options, setOptions] = useState<ASOption[]>([]);
    useEffect(() => {
      const doSearch = async () => {
        setIsLoading();

        const res = await onSearch(value ?? '');
        setOptions(res);

        unsetIsLoading();
      };

      void doSearch();
    }, [onSearch, setIsLoading, unsetIsLoading, value]);

    return (
      <Command
        ref={dropdownRef}
        {...commandProps}
        className={cn(
          'h-auto overflow-visible bg-transparent',
          commandProps?.className,
        )}
      >
        <div
          className={cn(
            'min-h-10 rounded-md border border-input text-base md:text-sm ring-offset-background focus-within:ring-1 focus-within:ring-ring focus-within:ring-offset-0',
            className,
          )}
          onClick={() => {
            if (disabled) return;
            inputRef.current?.focus();
          }}
        >
          <CommandPrimitive.Input
            {...inputProps}
            className={cn(
              'flex-1 bg-transparent outline-hidden placeholder:text-muted-foreground px-3 py-2 w-full',
              inputProps?.className,
            )}
            ref={inputRef}
            value={value}
            draggable
            onValueChange={onChange}
            placeholder={placeholder}
            onBlur={(event) => {
              if (!onScrollbar) {
                setOpen(false);
              }
              inputProps?.onBlur?.(event);
            }}
            onFocus={(event) => {
              setOpen(true);
              inputProps?.onFocus?.(event);
            }}
          />
        </div>
        <div className="relative">
          {open && (
            <CommandList
              className="absolute top-1 z-10 w-full rounded-md border bg-popover text-popover-foreground shadow-md outline-hidden animate-in"
              onMouseLeave={() => {
                setOnScrollbar(false);
              }}
              onMouseEnter={() => {
                setOnScrollbar(true);
              }}
              onMouseUp={() => {
                inputRef.current?.focus();
              }}
            >
              {isLoading ? (
                <>Loading...</>
              ) : !options.length ? (
                <>No results</>
              ) : (
                <CommandGroup className="h-full overflow-auto">
                  <>
                    {options.map((option) => (
                      <CommandItem
                        key={option}
                        value={option}
                        onMouseDown={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                        }}
                        onSelect={() => {
                          onChange(option);
                        }}
                      >
                        {option}
                      </CommandItem>
                    ))}
                  </>
                </CommandGroup>
              )}
            </CommandList>
          )}
        </div>
      </Command>
    );
  },
);
AsyncSelect.displayName = 'AsyncSelect';
