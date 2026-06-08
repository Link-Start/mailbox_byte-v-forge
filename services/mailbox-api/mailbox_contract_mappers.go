package main

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

func publicOperationAction(value string) mailboxv1.MailboxOperationAction {
	switch value {
	case operationActionRegisterMailbox:
		return mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_REGISTER_MAILBOX
	case operationActionMailboxOAuth:
		return mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_MAILBOX_OAUTH
	case operationActionFetchInboxes:
		return mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_FETCH_INBOXES
	default:
		return mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_UNSPECIFIED
	}
}

func operationActionValue(action mailboxv1.MailboxOperationAction) string {
	switch action {
	case mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_REGISTER_MAILBOX:
		return operationActionRegisterMailbox
	case mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_MAILBOX_OAUTH:
		return operationActionMailboxOAuth
	case mailboxv1.MailboxOperationAction_MAILBOX_OPERATION_ACTION_FETCH_INBOXES:
		return operationActionFetchInboxes
	default:
		return ""
	}
}

func publicOperationStatus(value string) mailboxv1.MailboxOperationStatus {
	switch value {
	case operationStatusCreated:
		return mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_CREATED
	case operationStatusRunning:
		return mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_RUNNING
	case operationStatusSucceeded:
		return mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_SUCCEEDED
	case operationStatusFailed:
		return mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_FAILED
	default:
		return mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_UNSPECIFIED
	}
}

func operationStatusValue(status mailboxv1.MailboxOperationStatus) string {
	switch status {
	case mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_CREATED:
		return operationStatusCreated
	case mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_RUNNING:
		return operationStatusRunning
	case mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_SUCCEEDED:
		return operationStatusSucceeded
	case mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_FAILED:
		return operationStatusFailed
	default:
		return ""
	}
}

func operationStatusFinal(status mailboxv1.MailboxOperationStatus) bool {
	return status == mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_SUCCEEDED ||
		status == mailboxv1.MailboxOperationStatus_MAILBOX_OPERATION_STATUS_FAILED
}
