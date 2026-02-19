import { Modal } from './Modal';

interface ConfirmDialogProps {
  title?: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onClose: () => void;
}

export function ConfirmDialog({
  title = 'Are you sure?',
  message,
  confirmLabel = 'Delete',
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  return (
    <Modal
      title={title}
      onClose={onClose}
      footer={
        <>
          <button className="btn btn-ghost" onClick={onClose}>
            Cancel
          </button>
          <button className="btn btn-danger" onClick={onConfirm}>
            {confirmLabel}
          </button>
        </>
      }
    >
      <p style={{ fontWeight: 500, lineHeight: 1.6 }}>{message}</p>
    </Modal>
  );
}
