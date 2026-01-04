<div id="lxdapi-panel-content" style="display:none;">
{if $success_msg}
<div class="alert alert-success"><i class="fas fa-check-circle"></i> {$success_msg}</div>
{/if}
{if $error_msg}
<div class="alert alert-danger"><i class="fas fa-exclamation-triangle"></i> {$error_msg}</div>
{/if}

{if $terminated}
<div class="alert alert-warning">
    <h4 class="alert-heading">产品已销毁</h4>
    <p>该容器及其所有数据已被永久删除。如果您需要重新使用，请重新购买产品。</p>
</div>
{else}
    {if !$error_msg}
    <div class="card mb-3" style="border:none;box-shadow:0 .15rem 1.75rem 0 rgba(58,59,69,.15)">
        <div class="card-header py-2" style="background:#f8f9fc;border-bottom:1px solid #e3e6f0">
            <h6 class="m-0 font-weight-bold" style="color:#1e40af"><i class="fas fa-external-link-alt mr-2"></i>容器面板</h6>
        </div>
        <div class="card-body py-3">
            <div class="d-flex justify-content-between align-items-center">
                <div>
                    <a href="{$jump_url}" target="_blank" class="btn" style="background:#2563eb;border-color:#2563eb;color:#fff"><i class="fas fa-external-link-alt"></i> 进入面板</a>
                    <span class="ml-3 text-muted">点击按钮将在新窗口打开容器管理面板</span>
                </div>
                <div>
                    <form method="post" action="" onsubmit="return confirmDestroy();">
                        <input type="hidden" name="customAction" value="terminate_container">
                        <button type="submit" class="btn btn-danger"><i class="fas fa-trash-alt"></i> 销毁容器</button>
                    </form>
                </div>
            </div>
        </div>
    </div>
    {if $iframe_url}
    <div class="card" style="border:none;box-shadow:0 .15rem 1.75rem 0 rgba(58,59,69,.15)">
        <div class="card-header py-2" style="background:#f8f9fc;border-bottom:1px solid #e3e6f0">
            <h6 class="m-0 font-weight-bold" style="color:#1e40af"><i class="fas fa-desktop mr-2"></i>容器管理</h6>
        </div>
        <div class="card-body p-0">
            <iframe src="{$iframe_url}" style="width:100%;height:100vh;border:none;display:block;" frameborder="0"></iframe>
        </div>
    </div>
    {/if}
    {/if}
{/if}
</div>

<script>
function confirmDestroy() {
    return confirm("警告：销毁容器将永久删除所有数据且无法恢复！\n您确定要继续吗？");
}

$(document).ready(function(){
    var overviewList = $('[menuitemname="Service Details Overview"] .list-group').first();
    var tabContent = $('.tab-content').first();
    if (overviewList.length && tabContent.length) {
        overviewList.append('<a menuitemname="Management" href="#tabManage" class="list-group-item list-group-item-action" data-toggle="list" role="tab"><div class="sidebar-menu-item-wrapper"><div class="sidebar-menu-item-label">管理</div></div></a>');
        tabContent.append('<div class="tab-pane fade" id="tabManage" role="tabpanel">' + $('#lxdapi-panel-content').html() + '</div>');
        $('#lxdapi-panel-content').remove();
        $('[menuitemname="Service Details Actions"]').hide();
    }
});
</script>
