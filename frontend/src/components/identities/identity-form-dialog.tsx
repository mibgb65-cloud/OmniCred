import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, LoaderCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { Controller, useForm, type UseFormRegisterReturn } from "react-hook-form";
import { z } from "zod";

import type { IdentityProfile, IdentityProfileInput } from "@/api/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const optionalEmail = z.string().trim().max(320, "邮箱不能超过 320 个字符").refine(
  (value) => !value || z.string().email().safeParse(value).success,
  "请输入有效的邮箱地址",
);

const schema = z.object({
  country: z.string().trim().min(1, "请输入国家或地区").max(256, "国家或地区名称过长"),
  full_name: z.string().trim().min(1, "请输入完整姓名").max(256, "完整姓名过长"),
  localized_name: z.string().trim().max(256, "中文名或本地语言姓名过长"),
  first_name: z.string().trim().max(256, "名过长"),
  middle_name: z.string().trim().max(256, "中间名过长"),
  last_name: z.string().trim().max(256, "姓过长"),
  gender: z.enum(["", "male", "female", "other"]),
  birth_date: z.string().refine((value) => !value || /^\d{4}-\d{2}-\d{2}$/.test(value), "请选择有效的出生日期"),
  street_address: z.string().trim().max(2048, "街道地址过长"),
  city: z.string().trim().max(256, "城市名称过长"),
  region: z.string().trim().max(256, "地区或省份名称过长"),
  postal_code: z.string().trim().max(32, "邮编过长"),
  phone: z.string().trim().max(64, "电话号码过长"),
  email: optionalEmail,
  password: z.string().max(16384, "密码过长"),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  country: "",
  full_name: "",
  localized_name: "",
  first_name: "",
  middle_name: "",
  last_name: "",
  gender: "",
  birth_date: "",
  street_address: "",
  city: "",
  region: "",
  postal_code: "",
  phone: "",
  email: "",
  password: "",
};

interface IdentityFormDialogProps {
  open: boolean;
  profile: IdentityProfile | null;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: IdentityProfileInput) => Promise<void>;
}

interface TextFieldProps {
  id: string;
  label: string;
  registration: UseFormRegisterReturn;
  error?: string;
  helper?: string;
  placeholder?: string;
  required?: boolean;
  type?: "text" | "date" | "email" | "tel";
  autoComplete?: string;
}

function TextField(props: TextFieldProps) {
  const descriptionID = props.error ? `${props.id}-error` : props.helper ? `${props.id}-helper` : undefined;
  return (
    <div className="space-y-2">
      <Label htmlFor={props.id}>
        {props.label} {props.required && <span className="text-destructive">*</span>}
      </Label>
      <Input
        id={props.id}
        type={props.type ?? "text"}
        placeholder={props.placeholder}
        autoComplete={props.autoComplete}
        aria-invalid={Boolean(props.error)}
        aria-describedby={descriptionID}
        {...props.registration}
      />
      {props.error
        ? <p id={`${props.id}-error`} className="text-xs font-medium text-destructive" role="alert">{props.error}</p>
        : props.helper && <p id={`${props.id}-helper`} className="text-xs leading-5 text-muted-foreground">{props.helper}</p>}
    </div>
  );
}

export function IdentityFormDialog(props: IdentityFormDialogProps) {
  const [showPassword, setShowPassword] = useState(false);
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const editing = props.profile !== null;

  useEffect(() => {
    if (!props.open) return;
    form.reset(props.profile ? {
      country: props.profile.country,
      full_name: props.profile.full_name,
      localized_name: props.profile.localized_name,
      first_name: props.profile.first_name,
      middle_name: props.profile.middle_name,
      last_name: props.profile.last_name,
      gender: props.profile.gender as FormValues["gender"],
      birth_date: props.profile.birth_date,
      street_address: props.profile.street_address,
      city: props.profile.city,
      region: props.profile.region,
      postal_code: props.profile.postal_code,
      phone: props.profile.phone,
      email: props.profile.email,
      password: props.profile.password,
    } : emptyValues);
    setShowPassword(false);
  }, [form, props.open, props.profile]);

  const busy = props.pending || form.formState.isSubmitting;
  const errors = form.formState.errors;

  return (
    <Dialog open={props.open} onOpenChange={(open) => !busy && props.onOpenChange(open)}>
      <DialogContent className="grid max-h-[92dvh] max-w-3xl grid-rows-[auto_minmax(0,1fr)] gap-0 overflow-hidden p-0 sm:p-0">
        <DialogHeader className="shrink-0 border-b border-border bg-card px-6 py-5 pr-14 sm:px-7 sm:pr-14">
          <DialogTitle>{editing ? "编辑身份资料" : "新增身份资料"}</DialogTitle>
          <DialogDescription>
            通用字段适用于不同国家；请按证件原文填写，不适用的项目可以留空。
          </DialogDescription>
        </DialogHeader>

        <form className="flex min-h-0 flex-col" onSubmit={form.handleSubmit(props.onSubmit)} noValidate>
          <div className="scrollbar-hidden min-h-0 flex-1 space-y-6 overflow-y-auto overscroll-contain px-6 py-5 sm:px-7">
            <fieldset className="space-y-4" disabled={busy}>
              <legend className="text-sm font-bold text-foreground">基本信息</legend>
              <div className="grid gap-4 sm:grid-cols-2">
                <TextField id="identity-country" label="国家 / 地区" required placeholder="例如：Philippines" autoComplete="country-name" registration={form.register("country")} error={errors.country?.message} />
                <TextField id="identity-full-name" label="完整姓名（Full Name）" required placeholder="例如：Angelo Santos" autoComplete="name" registration={form.register("full_name")} error={errors.full_name?.message} />
                <TextField id="identity-localized-name" label="中文名 / 本地语言姓名" placeholder="例如：安杰洛·桑托斯" registration={form.register("localized_name")} error={errors.localized_name?.message} />
                <TextField id="identity-first-name" label="名（First Name）" placeholder="例如：Angelo" autoComplete="given-name" registration={form.register("first_name")} error={errors.first_name?.message} />
                <TextField id="identity-middle-name" label="中间名（Middle Name）" placeholder="例如：Reyes" autoComplete="additional-name" registration={form.register("middle_name")} error={errors.middle_name?.message} helper="菲律宾资料中通常填写母姓；其他国家请按证件原文填写。" />
                <TextField id="identity-last-name" label="姓（Last Name）" placeholder="例如：Santos" autoComplete="family-name" registration={form.register("last_name")} error={errors.last_name?.message} />

                <Controller
                  name="gender"
                  control={form.control}
                  render={({ field, fieldState }) => (
                    <div className="space-y-2">
                      <Label htmlFor="identity-gender">性别</Label>
                      <Select value={field.value || "not-specified"} onValueChange={(value) => field.onChange(value === "not-specified" ? "" : value)} disabled={busy}>
                        <SelectTrigger ref={field.ref} id="identity-gender" onBlur={field.onBlur} aria-invalid={fieldState.invalid}>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="not-specified">未填写</SelectItem>
                          <SelectItem value="male">Male（男）</SelectItem>
                          <SelectItem value="female">Female（女）</SelectItem>
                          <SelectItem value="other">Other（其他）</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                />
                <TextField id="identity-birth-date" label="出生日期" type="date" autoComplete="bday" registration={form.register("birth_date")} error={errors.birth_date?.message} />
              </div>
            </fieldset>

            <fieldset className="space-y-4 border-t border-border pt-5" disabled={busy}>
              <legend className="text-sm font-bold text-foreground">地址</legend>
              <TextField id="identity-street" label="街道地址" placeholder="Unit 402, Sunshine Court, Aurora Blvd, Cubao" autoComplete="street-address" registration={form.register("street_address")} error={errors.street_address?.message} />
              <div className="grid gap-4 sm:grid-cols-3">
                <TextField id="identity-city" label="城市" placeholder="Quezon City" autoComplete="address-level2" registration={form.register("city")} error={errors.city?.message} />
                <TextField id="identity-region" label="地区 / 省份" placeholder="Metro Manila" autoComplete="address-level1" registration={form.register("region")} error={errors.region?.message} />
                <TextField id="identity-postal-code" label="邮编（ZIP Code）" placeholder="1109" autoComplete="postal-code" registration={form.register("postal_code")} error={errors.postal_code?.message} />
              </div>
            </fieldset>

            <fieldset className="space-y-4 border-t border-border pt-5" disabled={busy}>
              <legend className="text-sm font-bold text-foreground">联系与登录</legend>
              <div className="grid gap-4 sm:grid-cols-2">
                <TextField id="identity-phone" label="电话" type="tel" placeholder="+63 (917) 482-9301" autoComplete="tel" registration={form.register("phone")} error={errors.phone?.message} />
                <TextField id="identity-email" label="邮箱" type="email" placeholder="name@example.com" autoComplete="email" registration={form.register("email")} error={errors.email?.message} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="identity-password">密码</Label>
                <div className="relative">
                  <Input
                    id="identity-password"
                    type={showPassword ? "text" : "password"}
                    autoComplete="new-password"
                    className="pr-12"
                    aria-invalid={Boolean(errors.password)}
                    aria-describedby={errors.password ? "identity-password-error" : "identity-password-helper"}
                    {...form.register("password")}
                  />
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => setShowPassword((value) => !value)} className="absolute right-1.5 top-1/2 -translate-y-1/2" aria-label={showPassword ? "隐藏身份资料密码" : "显示身份资料密码"} aria-pressed={showPassword}>
                    {showPassword ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}
                  </Button>
                </div>
                {errors.password
                  ? <p id="identity-password-error" className="text-xs font-medium text-destructive" role="alert">{errors.password.message}</p>
                  : <p id="identity-password-helper" className="text-xs leading-5 text-muted-foreground">可选。当前版本会直接写入本机数据库；不需要时请留空。</p>}
              </div>
            </fieldset>
          </div>

          <DialogFooter className="shrink-0 border-t border-border bg-card px-6 py-4 sm:px-7">
            <Button type="button" variant="outline" onClick={() => props.onOpenChange(false)} disabled={busy}>取消</Button>
            <Button type="submit" disabled={busy}>
              {busy && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
              {editing ? "保存修改" : "保存身份资料"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
