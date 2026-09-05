import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

function daysInMonth(year: string, month: string) {
  const numericMonth = Number(month);
  if (numericMonth === 2) {
    if (!/^\d{4}$/.test(year)) return 29;
    const numericYear = Number(year);
    return numericYear % 400 === 0 || (numericYear % 4 === 0 && numericYear % 100 !== 0) ? 29 : 28;
  }
  return [4, 6, 9, 11].includes(numericMonth) ? 30 : 31;
}

export function isValidBirthDate(value: string) {
  if (!value) return true;
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) return false;
  const [, year, month, day] = match;
  const numericMonth = Number(month);
  const numericDay = Number(day);
  return Number(year) > 0
    && numericMonth >= 1
    && numericMonth <= 12
    && numericDay >= 1
    && numericDay <= daysInMonth(year, month);
}

interface BirthDateFieldProps {
  value: string;
  error?: string;
  inputRef: (instance: HTMLInputElement | null) => void;
  onBlur: () => void;
  onChange: (value: string) => void;
}

const months = Array.from({ length: 12 }, (_, index) => String(index + 1).padStart(2, "0"));

function birthDateParts(value: string) {
  const [year = "", month = "", day = ""] = value.split("-");
  return { year, month, day };
}

function birthDateValue(year: string, month: string, day: string) {
  return year || month || day ? `${year}-${month}-${day}` : "";
}

export function BirthDateField(props: BirthDateFieldProps) {
  const { year, month, day } = birthDateParts(props.value);
  const maximumDay = daysInMonth(year, month);
  const days = Array.from({ length: maximumDay }, (_, index) => String(index + 1).padStart(2, "0"));
  const descriptionID = props.error ? "identity-birth-date-error" : "identity-birth-date-helper";

  const updateYear = (nextYear: string) => {
    const sanitizedYear = nextYear.replace(/\D/g, "").slice(0, 4);
    const nextDay = Number(day) > daysInMonth(sanitizedYear, month) ? "" : day;
    props.onChange(birthDateValue(sanitizedYear, month, nextDay));
  };

  const updateMonth = (nextMonth: string) => {
    const normalizedMonth = nextMonth === "not-specified" ? "" : nextMonth;
    const nextDay = Number(day) > daysInMonth(year, normalizedMonth) ? "" : day;
    props.onChange(birthDateValue(year, normalizedMonth, nextDay));
  };

  return (
    <div className="space-y-2">
      <Label id="identity-birth-date-label">出生日期</Label>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-[minmax(5rem,1fr)_5rem_5rem]" role="group" aria-labelledby="identity-birth-date-label">
        <Input
          ref={props.inputRef}
          id="identity-birth-year"
          className="col-span-2 text-center tabular-nums sm:col-span-1"
          type="text"
          inputMode="numeric"
          maxLength={4}
          autoComplete="bday-year"
          value={year}
          placeholder="年份"
          aria-label="出生年份"
          aria-invalid={Boolean(props.error)}
          aria-describedby={descriptionID}
          onBlur={props.onBlur}
          onChange={(event) => updateYear(event.target.value)}
        />
        <Select value={month || "not-specified"} onValueChange={updateMonth}>
          <SelectTrigger id="identity-birth-month" aria-label="出生月份" aria-invalid={Boolean(props.error)} aria-describedby={descriptionID} onBlur={props.onBlur}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="not-specified">月</SelectItem>
            {months.map((option) => <SelectItem key={option} value={option}>{Number(option)} 月</SelectItem>)}
          </SelectContent>
        </Select>
        <Select
          value={day || "not-specified"}
          onValueChange={(nextDay) => props.onChange(birthDateValue(year, month, nextDay === "not-specified" ? "" : nextDay))}
        >
          <SelectTrigger id="identity-birth-day" aria-label="出生日" aria-invalid={Boolean(props.error)} aria-describedby={descriptionID} onBlur={props.onBlur}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="not-specified">日</SelectItem>
            {days.map((option) => <SelectItem key={option} value={option}>{Number(option)} 日</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
      {props.error
        ? <p id="identity-birth-date-error" className="text-xs font-medium text-destructive" role="alert">{props.error}</p>
        : <p id="identity-birth-date-helper" className="text-xs leading-5 text-muted-foreground">按年、月、日填写，年份限 4 位数字。</p>}
    </div>
  );
}
