# run-monitor-chart

Clone the repo `git clone https://github.com/sjenning/run-monitor-chart`

Set your `KUBECONFIG` appropriately for `system:admin` user.

Capture cluster activity with `openshift-tests run-monitor | tee run-monitor.log`.

View clusteroperator condition transistions with `./show-chart.sh run-monitor.log`.

To view an example, `./show-chart.sh run-monitor-example.log`