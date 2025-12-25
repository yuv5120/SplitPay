import * as React from "react";
import { Check } from "lucide-react";
import { cn } from "../../lib/utils";

export const Checkbox = React.forwardRef(
  ({ className, checked, onCheckedChange, ...props }, ref) => {
    return (
      <button
        type="button"
        role="checkbox"
        aria-checked={checked}
        ref={ref}
        className={cn(
          "peer h-5 w-5 shrink-0 rounded border-2 border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all",
          checked ? "bg-primary text-primary-foreground" : "bg-background",
          className
        )}
        onClick={() => onCheckedChange(!checked)}
        {...props}
      >
        {checked && <Check className="h-4 w-4" />}
      </button>
    );
  }
);

Checkbox.displayName = "Checkbox";