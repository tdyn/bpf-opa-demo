/*
Copyright The Kubeshield Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

func lookupFillerID(event string) (int, bool) {
	for i, name := range fillerNames {
		if name == event {
			return i, true
		}
	}
	return -1, false
}

var fillerNames []string = []string{
	"sys_autofill",
	"sys_generic",
	"sys_empty",
	"sys_single",
	"sys_single_x",
	"sys_open_x",
	"sys_read_x",
	"sys_write_x",
	"sys_execve_e",
	"proc_startupdate",
	"proc_startupdate_2",
	"proc_startupdate_3",
	"sys_socketpair_x",
	"sys_setsockopt_x",
	"sys_getsockopt_x",
	"sys_connect_x",
	"sys_accept4_e",
	"sys_accept_x",
	"sys_send_e",
	"sys_send_x",
	"sys_sendto_e",
	"sys_sendmsg_e",
	"sys_sendmsg_x",
	"sys_recv_x",
	"sys_recvfrom_x",
	"sys_recvmsg_x",
	"sys_recvmsg_x_2",
	"sys_shutdown_e",
	"sys_creat_x",
	"sys_pipe_x",
	"sys_eventfd_e",
	"sys_futex_e",
	"sys_lseek_e",
	"sys_llseek_e",
	"sys_socket_bind_x",
	"sys_poll_e",
	"sys_poll_x",
	"sys_pread64_e",
	"sys_preadv64_e",
	"sys_writev_e",
	"sys_pwrite64_e",
	"sys_readv_preadv_x",
	"sys_writev_pwritev_x",
	"sys_pwritev_e",
	"sys_nanosleep_e",
	"sys_getrlimit_setrlimit_e",
	"sys_getrlimit_setrlrimit_x",
	"sys_prlimit_e",
	"sys_prlimit_x",
	"sched_switch_e",
	"sched_drop",
	"sys_fcntl_e",
	"sys_ptrace_e",
	"sys_ptrace_x",
	"sys_mmap_e",
	"sys_brk_munmap_mmap_x",
	"sys_renameat_x",
	"sys_symlinkat_x",
	"sys_procexit_e",
	"sys_sendfile_e",
	"sys_sendfile_x",
	"sys_quotactl_e",
	"sys_quotactl_x",
	"sys_sysdigevent_e",
	"sys_getresuid_and_gid_x",
	"sys_signaldeliver_e",
	"sys_pagefault_e",
	"sys_setns_e",
	"sys_unshare_e",
	"sys_flock_e",
	"cpu_hotplug_e",
	"sys_semop_x",
	"sys_semget_e",
	"sys_semctl_e",
	"sys_ppoll_e",
	"sys_mount_e",
	"sys_access_e",
	"sys_socket_x",
	"sys_bpf_x",
	"sys_unlinkat_x",
	"sys_fchmodat_x",
	"sys_chmod_x",
	"sys_fchmod_x",
	"sys_mkdirat_x",
	"sys_openat_x",
	"sys_linkat_x",
	"terminate_filler",
}
