import { useMutation } from '@tanstack/react-query';
import { ArrowBigRightDash } from 'lucide-react';
import React, {
  ChangeEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from 'react';

import { CompleteRecipeMutationOption } from '@/api/completions';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { VoidCB } from '@/types';
import { CompleteRecipeResponse } from '@/types/completions';

type OnImportFinsh = (
  recipe: CompleteRecipeResponse & {
    url: string;
  },
) => void;
const FakeProgress: React.FC<{ isLoading: boolean }> = ({ isLoading }) => {
  const [progress, setProgress] = useState(0);
  const speedFactor = 0.25;

  useEffect(() => {
    if (!isLoading) {
      setProgress(100);

      return;
    }
    // Reset progress when loading starts
    setProgress(0);

    // Simulate progress with non-linear increments
    const simulateProgress = () => {
      setProgress((currentProgress) => {
        if (currentProgress >= 100) return 100;

        // Slow down progress as it gets closer to 90%
        const remaining = 100 - currentProgress;
        const increment =
          remaining > 30
            ? Math.random() * 15 * speedFactor
            : Math.random() * 3 * speedFactor;

        return Math.min(currentProgress + increment, 90);
      });
    };

    // Update progress at random intervals
    const intervalId = setInterval(() => {
      simulateProgress();
    }, 200 / speedFactor);

    return () => {
      clearInterval(intervalId);
    };
  }, [isLoading, speedFactor]);

  return (
    <div>
      <Progress
        className="transition-all duration-500 ease-out"
        value={progress}
      />
    </div>
  );
};

const ProgressStep: React.FC<{ data?: CompleteRecipeResponse }> = ({
  data,
}) => (
  <>
    <DialogHeader>
      <DialogTitle>Processing recipe</DialogTitle>
    </DialogHeader>
    <FakeProgress isLoading={!data} />
  </>
);

const InputStep: React.FC<{
  error?: string;
  onSubmit: (url: string) => void;
}> = ({ error, onSubmit }) => {
  const [isValid, setIsValid] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleInputChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      setIsValid(event.target.validity.valid);
    },
    [],
  );

  const handleSubmit = useCallback(() => {
    onSubmit(inputRef.current?.value ?? '');
  }, [onSubmit]);

  return (
    <>
      <DialogHeader>
        <DialogTitle aria-description="">Enter URL</DialogTitle>
      </DialogHeader>
      <div className="flex items-center space-x-2">
        <Input
          id="url"
          placeholder="Enter url"
          type="url"
          onChange={handleInputChange}
          onReset={handleInputChange}
          ref={inputRef}
          required
        />
        <Button
          type="submit"
          className="px-3"
          size="sm"
          disabled={!isValid}
          onClick={handleSubmit}
        >
          <ArrowBigRightDash />
        </Button>
      </div>
      {error && (
        <p className="text-[0.8rem] font-medium text-destructive">
          An error occurred: {error}
        </p>
      )}
    </>
  );
};

export const ImportFromURLDialog: React.FC<{
  onFinish: OnImportFinsh;
}> = ({ onFinish }) => {
  const { mutate, data, error, isPending } = useMutation({
    ...CompleteRecipeMutationOption,
  });

  const handleSubmit = useCallback(
    (url: string) => {
      mutate(
        { url },
        {
          onSuccess: (data) => {
            onFinish({ ...data, url });
          },
        },
      );
    },
    [mutate, onFinish],
  );

  if (!isPending && !data) {
    return <InputStep onSubmit={handleSubmit} error={error?.message} />;
  }

  return <ProgressStep data={data} />;
};

export const ImportFromURL: React.FC<{
  isOpen: boolean;
  close: VoidCB;
  onFinish: OnImportFinsh;
}> = ({ isOpen, close, onFinish }) => {
  const onOpenChange = useCallback(
    (state: boolean) => {
      if (!state) {
        close();
      }
    },
    [close],
  );

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]" aria-describedby={undefined}>
        <ImportFromURLDialog onFinish={onFinish} />
      </DialogContent>
    </Dialog>
  );
};
