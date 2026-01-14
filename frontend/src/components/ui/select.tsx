import * as React from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "../../lib/utils";

interface SelectProps {
  value: string;
  onValueChange: (value: string) => void;
  children: React.ReactNode;
}

export const Select: React.FC<SelectProps> = ({ value, onValueChange, children }) => {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={ref} className="relative">
      {React.Children.map(children, (child: any) => {
        if (child?.type === SelectTrigger) {
          return React.cloneElement(child, { open, setOpen, value });
        }
        if (child?.type === SelectContent) {
          return React.cloneElement(child, { open, onValueChange, setOpen });
        }
        return child;
      })}
    </div>
  );
};

interface SelectTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  open?: boolean;
  setOpen?: (open: boolean) => void;
  value?: string;
  placeholder?: string;
}

export const SelectTrigger = React.forwardRef<HTMLButtonElement, SelectTriggerProps>(
  ({ className, children, open, setOpen, value, placeholder }, ref) => {
    let displayText = placeholder;
    
    React.Children.forEach(children, (child: any) => {
      if (child && child.type === SelectValue) {
        displayText = child.props.children || child.props.placeholder;
      }
    });

    return (
      <button
        ref={ref}
        type="button"
        onClick={() => setOpen?.(!open)}
        className={cn(
          "flex h-10 w-full items-center justify-between rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all",
          className
        )}
      >
        <span className={!value ? "text-muted-foreground" : ""}>
          {value ? displayText : (placeholder || "Select an option")}
        </span>
        <ChevronDown className="h-4 w-4 opacity-50" />
      </button>
    );
  }
);
SelectTrigger.displayName = "SelectTrigger";

interface SelectValueProps {
  placeholder?: string;
  children?: React.ReactNode;
}

export const SelectValue: React.FC<SelectValueProps> = ({ placeholder, children }) => {
  return <>{children || placeholder}</>;
};

interface SelectContentProps extends React.HTMLAttributes<HTMLDivElement> {
  open?: boolean;
  onValueChange?: (value: string) => void;
  setOpen?: (open: boolean) => void;
}

export const SelectContent = React.forwardRef<HTMLDivElement, SelectContentProps>(
  ({ className, children, open, onValueChange, setOpen }, ref) => {
    if (!open) return null;

    return (
      <div
        ref={ref}
        className={cn(
          "absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border bg-popover text-popover-foreground shadow-lg animate-fade-in",
          className
        )}
      >
        {React.Children.map(children, (child: any) => {
          if (child?.type === SelectItem) {
            return React.cloneElement(child, { onValueChange, setOpen });
          }
          return child;
        })}
      </div>
    );
  }
);
SelectContent.displayName = "SelectContent";

interface SelectItemProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string;
  onValueChange?: (value: string) => void;
  setOpen?: (open: boolean) => void;
}

export const SelectItem = React.forwardRef<HTMLDivElement, SelectItemProps>(
  ({ className, children, value, onValueChange, setOpen }, ref) => (
    <div
      ref={ref}
      onClick={() => {
        onValueChange?.(value);
        setOpen?.(false);
      }}
      className={cn(
        "relative flex cursor-pointer select-none items-center rounded-sm py-2 px-3 text-sm outline-none hover:bg-accent hover:text-accent-foreground transition-colors",
        className
      )}
    >
      {children}
    </div>
  )
);
SelectItem.displayName = "SelectItem";
