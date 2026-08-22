import { LoaderCircle, TriangleAlert } from "lucide-react";

import type { Credential } from "@/api/types";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface DeleteCredentialDialogProps {
  credential: Credential | null;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function DeleteCredentialDialog(props: DeleteCredentialDialogProps) {
  return (
    <AlertDialog open={props.credential !== null} onOpenChange={(open) => !props.pending && props.onOpenChange(open)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <div className="grid size-11 place-items-center rounded-xl bg-destructive/10 text-destructive">
            <TriangleAlert className="size-5" aria-hidden="true" />
          </div>
          <AlertDialogTitle>删除这条账号记录？</AlertDialogTitle>
          <AlertDialogDescription>
            {props.credential ? `“${props.credential.account}” 将从本地数据库中永久删除。此操作无法撤销。` : "此操作无法撤销。"}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={props.pending}>取消</AlertDialogCancel>
          <AlertDialogAction onClick={props.onConfirm} disabled={props.pending}>
            {props.pending && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
            确认删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
